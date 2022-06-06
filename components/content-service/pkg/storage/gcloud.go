// Copyright (c) 2020 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package storage

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	gcpstorage "cloud.google.com/go/storage"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/opentracing/opentracing-go"
	"golang.org/x/oauth2/google"
	"golang.org/x/xerrors"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/gitpod-io/gitpod/common-go/log"
	"github.com/gitpod-io/gitpod/common-go/tracing"
	config "github.com/gitpod-io/gitpod/content-service/api/config"
	"github.com/gitpod-io/gitpod/content-service/pkg/archive"
)

var _ DirectAccess = &DirectGCPStorage{}

var validateExistsInFilesystem = validation.By(func(o interface{}) error {
	s, ok := o.(string)
	if !ok {
		return xerrors.Errorf("field should be string")
	}

	if s == "" {
		// don't make this field required
		return nil
	}

	_, err := os.Stat(s)
	return err
})

// Validate checks if the GCloud storage GCPconfig is valid
func ValidateGCPConfig(c *config.GCPConfig) error {
	return validation.ValidateStruct(c,
		validation.Field(&c.CredentialsFile, validateExistsInFilesystem),
		validation.Field(&c.Region, validation.Required),
		validation.Field(&c.Project, validation.Required),
	)
}

// newDirectGCPAccess provides direct access to the remote storage system
func newDirectGCPAccess(cfg config.GCPConfig, stage config.Stage) (*DirectGCPStorage, error) {
	if err := ValidateGCPConfig(&cfg); err != nil {
		return nil, err
	}

	return &DirectGCPStorage{
		Stage:     stage,
		GCPConfig: cfg,
	}, nil
}

// DirectGCPStorage stores data in Google Cloud buckets, following a particular naming scheme
type DirectGCPStorage struct {
	Username      string
	WorkspaceName string
	InstanceID    string
	GCPConfig     config.GCPConfig
	Stage         config.Stage

	client *gcpstorage.Client

	// ObjectAccess just exists so that we can swap out the stream access during testing
	ObjectAccess func(ctx context.Context, btk, obj string) (io.ReadCloser, bool, error)
}

// Validate checks if the GCloud storage is GCPconfigured properly
func (rs *DirectGCPStorage) Validate() error {
	err := ValidateGCPConfig(&rs.GCPConfig)
	if err != nil {
		return err
	}

	return validation.ValidateStruct(rs,
		validation.Field(&rs.Username, validation.Required),
		validation.Field(&rs.WorkspaceName, validation.Required),
		validation.Field(&rs.Stage, validation.Required),
	)
}

// Init initializes the remote storage - call this before calling anything else on the interface
func (rs *DirectGCPStorage) Init(ctx context.Context, owner, workspace, instance string) (err error) {
	//nolint:ineffassign
	span, ctx := opentracing.StartSpanFromContext(ctx, "GCloudBucketRemotegcpStorage.Init")
	defer tracing.FinishSpan(span, &err)

	rs.Username = owner
	rs.WorkspaceName = workspace
	rs.InstanceID = instance

	// now that we have all the information complete, validate if we're good to go
	err = rs.Validate()
	if err != nil {
		return xerrors.Errorf("invalid GCloud remote storage GCPconfig: %w", err)
	}

	client, err := newGCPClient(ctx, rs.GCPConfig)
	if err != nil {
		return err
	}
	rs.client = client

	if rs.ObjectAccess == nil {
		rs.ObjectAccess = rs.defaultObjectAccess
	}

	return nil
}

// EnsureExists makes sure that the remote storage location exists and can be up- or downloaded from
func (rs *DirectGCPStorage) EnsureExists(ctx context.Context) (err error) {
	return gcpEnsureExists(ctx, rs.client, rs.bucketName(), rs.GCPConfig)
}

func gcpEnsureExists(ctx context.Context, client *gcpstorage.Client, bucketName string, gcpConfig config.GCPConfig) (err error) {
	//nolint:ineffassign
	span, ctx := opentracing.StartSpanFromContext(ctx, "GCloudBucketRemotegcpStorage.EnsureExists")
	defer tracing.FinishSpan(span, &err)

	if client == nil {
		return xerrors.Errorf("no gcloud client available - did you call Init()?")
	}

	hdl := client.Bucket(bucketName)
	_, err = hdl.Attrs(ctx)
	if err == nil {
		// bucket exists and everything is fine - we're done here
		return
	}
	if err != nil && err != gcpstorage.ErrBucketNotExist {
		return xerrors.Errorf("cannot ensure storage exists: %w", err)
	}

	log.WithField("bucketName", bucketName).Debug("Creating bucket")
	err = hdl.Create(ctx, gcpConfig.Project, &gcpstorage.BucketAttrs{
		Location: gcpConfig.Region,
	})
	if e, ok := err.(*googleapi.Error); ok && e.Code == http.StatusConflict && strings.Contains(strings.ToLower(e.Message), "you already own this bucket") {
		// Looks like we had a bucket creation race and lost.
		// That's ok - at least the bucket exists now and is still owned by us.
	} else if err != nil {
		return xerrors.Errorf("cannot create bucket: %w", err)
	}

	return nil
}

func (rs *DirectGCPStorage) defaultObjectAccess(ctx context.Context, bkt, obj string) (io.ReadCloser, bool, error) {
	if rs.client == nil {
		return nil, false, xerrors.Errorf("no gcloud client available - did you call Init()?")
	}

	objHandle := rs.client.Bucket(bkt).Object(obj)
	rc, err := objHandle.NewReader(ctx)
	if err != nil {
		return nil, false, err
	}

	return rc, false, nil
}

func (rs *DirectGCPStorage) download(ctx context.Context, destination string, bkt string, obj string, mappings []archive.IDMapping) (found bool, err error) {
	//nolint:ineffassign
	span, ctx := opentracing.StartSpanFromContext(ctx, "GCloudBucketRemotegcpStorage.Download")
	span.SetTag("gcsBkt", bkt)
	span.SetTag("gcsObj", obj)
	defer tracing.FinishSpan(span, &err)

	var rc io.ReadCloser
	t0 := time.Now()
	if os.Getenv("USE_GSUTIL_FOR_DOWNLOADS") == "true" {
		log.Info("Using gsutil to download GCS content")
		tmpDir, err := os.MkdirTemp("ws-daemon", "gcs-download-")
		if err != nil {
			return false, fmt.Errorf("cannot create temportal file: %w", err)
		}
		defer os.RemoveAll(tmpDir)

		args := fmt.Sprintf(`gcloud auth activate-service-account --key-file %v \
		&& gsutil \
		  -o "GSUtil:parallel_thread_count=1" \
		  -o "GSUtil:sliced_object_download_max_components=8" \
		  cp gs://%s %s`, rs.GCPConfig.CredentialsFile, filepath.Join(bkt, obj), tmpDir)

		cmd := exec.Command("/bin/bash", []string{"-c", args}...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.WithError(err).WithField("out", string(out)).Error("unexpected error downloading file from GCS using gsutil")
			return false, err
		}

		file := filepath.Base(bkt)
		rc, err = os.Open(filepath.Join(tmpDir, file))
		if err != nil {
			return false, err
		}

		log.WithField("object", filepath.Join(bkt, obj)).WithField("gsutil", args).WithField("pullTime", time.Since(t0)).Info("GCS content download time")
	} else {
		rc, _, err = rs.ObjectAccess(ctx, bkt, obj)
		if rc == nil {
			return false, err
		}
	}

	defer rc.Close()

	err = extractTarbal(ctx, destination, rc, mappings)
	if err != nil {
		return true, err
	}

	if err := rs.fixLegacyFilenames(ctx, destination); err != nil {
		return true, err
	}

	return true, nil
}

/* tar files produced by the previous sync process contain their workspace ID in the filenames.
 * This behavior is difficult for snapshot backups, thus ws-daemond does not do that. However,
 * we need to be able to handle the "old" tar files, hence this legacy mode. See #1559.
 */
func (rs *DirectGCPStorage) fixLegacyFilenames(ctx context.Context, destination string) (err error) {
	//nolint:staticcheck,ineffassign
	span, ctx := opentracing.StartSpanFromContext(ctx, "fixLegacyFilenames")
	defer tracing.FinishSpan(span, &err)

	legacyPath := filepath.Join(destination, rs.WorkspaceName)
	if fi, err := os.Stat(legacyPath); os.IsNotExist(err) {
		// legacy path does not exist, nothing to do here
		return nil
	} else if fi.IsDir() {
		log.WithField("destination", destination).WithField("legacyPath", legacyPath).Info("Handling legacy backup")
		/* legacy path exists and is a directory - move it's content and remove the legacy path.
		 *
		 * Using mv here is difficult as the wildcard expansion is done by the shell and not mv,
		 * thus we'd need to wrap the mv call in a sh call -> too many dependencies to the outside world.
		 */
		fis, err := os.ReadDir(legacyPath)
		if err != nil {
			return err
		}
		for _, fi := range fis {
			src := filepath.Join(legacyPath, fi.Name())
			dst := filepath.Join(destination, fi.Name())
			log.WithField("src", src).WithField("dst", dst).Debug("moving file")
			if err := os.Rename(src, dst); err != nil {
				return xerrors.Errorf("mv %s %s: %s", src, dst, err)
			}
		}

		if err := os.Remove(legacyPath); err != nil {
			return err
		}
	}

	return nil
}

// Download takes the latest state from the remote storage and downloads it to a local path
func (rs *DirectGCPStorage) Download(ctx context.Context, destination string, name string, mappings []archive.IDMapping) (bool, error) {
	return rs.download(ctx, destination, rs.bucketName(), rs.objectName(name), mappings)
}

// DownloadSnapshot downloads a snapshot. The snapshot name is expected to be one produced by Qualify
func (rs *DirectGCPStorage) DownloadSnapshot(ctx context.Context, destination string, name string, mappings []archive.IDMapping) (bool, error) {
	bkt, obj, err := ParseSnapshotName(name)
	if err != nil {
		return false, err
	}

	return rs.download(ctx, destination, bkt, obj, mappings)
}

// ParseSnapshotName parses the name of a snapshot into bucket and object
func ParseSnapshotName(name string) (bkt, obj string, err error) {
	segments := strings.Split(name, "@")
	if len(segments) != 2 {
		err = xerrors.Errorf("%s is not a valid GCloud remote storage FQN", name)
		return
	}

	obj = segments[0]
	bkt = segments[1]
	return
}

// ListObjects returns all objects found with the given prefix. Returns an empty list if the bucket does not exuist (yet).
func (rs *DirectGCPStorage) ListObjects(ctx context.Context, prefix string) (objects []string, err error) {
	bkt := rs.client.Bucket(rs.bucketName())
	_, err = bkt.Attrs(ctx)
	if errors.Is(err, gcpstorage.ErrBucketNotExist) {
		// bucket does not exist: nothing to list
		return nil, nil
	}
	if err != nil {
		return nil, xerrors.Errorf("cannot list objects: %w", err)
	}

	iter := bkt.Objects(ctx, &gcpstorage.Query{Prefix: prefix})
	var obj *gcpstorage.ObjectAttrs
	for obj, err = iter.Next(); obj != nil; obj, err = iter.Next() {
		objects = append(objects, obj.Name)
	}
	if err != iterator.Done && err != nil {
		return nil, xerrors.Errorf("cannot iterate list objects: %w", err)
	}

	return objects, nil
}

// Qualify fully qualifies a snapshot name so that it can be downloaded using DownloadSnapshot
func (rs *DirectGCPStorage) Qualify(name string) string {
	return fmt.Sprintf("%s@%s", rs.objectName(name), rs.bucketName())
}

// UploadInstance takes all files from a local location and uploads it to the per-instance remote storage
func (rs *DirectGCPStorage) UploadInstance(ctx context.Context, source string, name string, opts ...UploadOption) (bucket, object string, err error) {
	if rs.InstanceID == "" {
		return "", "", xerrors.Errorf("instanceID is required to comput object name")
	}
	return rs.Upload(ctx, source, InstanceObjectName(rs.InstanceID, name), opts...)
}

// Upload takes all files from a local location and uploads it to the remote storage
func (rs *DirectGCPStorage) Upload(ctx context.Context, source string, name string, opts ...UploadOption) (bucket, object string, err error) {
	//nolint:ineffassign
	span, ctx := opentracing.StartSpanFromContext(ctx, "GCloudBucketRemotegcpStorage.Upload")
	defer tracing.FinishSpan(span, &err)
	log := log.WithFields(log.OWI(rs.Username, rs.WorkspaceName, ""))

	options, err := GetUploadOptions(opts)
	if err != nil {
		err = xerrors.Errorf("cannot get options: %w", err)
		return
	}

	if rs.client == nil {
		err = xerrors.Errorf("no gcloud client available - did you call Init()?")
		return
	}

	// check if we have not yet exceeded the max number of backups
	if name != DefaultBackup {
		if err = rs.ensureBackupSlotAvailable(); err != nil {
			return
		}
	}

	sfn, err := os.Open(source)
	if err != nil {
		err = xerrors.Errorf("cannot open file for uploading: %w", err)
		return
	}
	defer sfn.Close()

	stat, err := sfn.Stat()
	if err != nil {
		return
	}

	totalSize := stat.Size()
	span.SetTag("totalSize", totalSize)

	bucket = rs.bucketName()
	bkt := rs.client.Bucket(bucket)

	object = rs.objectName(name)
	obj := bkt.Object(object)

	uploadSpan := opentracing.StartSpan("remote-upload", opentracing.ChildOf(span.Context()))
	uploadSpan.SetTag("bucket", bucket)
	uploadSpan.SetTag("obj", object)

	var firstBackup bool
	if _, e := obj.Attrs(ctx); e == gcpstorage.ErrObjectNotExist {
		firstBackup = true
	}

	var wg sync.WaitGroup
	var written int64

	wg.Add(1)

	go func() {
		defer wg.Done()

		wc := obj.NewWriter(ctx)
		wc.Metadata = options.Annotations
		wc.ContentType = options.ContentType
		// Increase chunk size for faster uploading
		wc.ChunkSize = googleapi.DefaultUploadChunkSize * 4

		written, err = io.Copy(wc, sfn)
		if err != nil {
			log.WithError(err).WithField("name", name).Error("Error while uploading file")
			return
		}

		// persist changes in GCS
		err = wc.Close()
		if err != nil {
			log.WithError(err).WithField("name", name).Error("Error while uploading file")
			return
		}
	}()

	wg.Wait()

	if written != totalSize {
		err = xerrors.Errorf("Wrote fewer bytes than it should have, %d instead of %d", written, totalSize)
		return
	}

	// maintain backup trail if we're asked to - we do this prior to overwriting the regular backup file
	// to make sure we're trailign the previous backup.
	if options.BackupTrail.Enabled && !firstBackup {
		err := rs.trailBackup(ctx, bkt, obj, options.BackupTrail.ThisBackupID, options.BackupTrail.TrailLength)
		if err != nil {
			log.WithError(err).Error("cannot maintain backup trail")
		}
	}

	uploadSpan.Finish()

	err = nil
	return
}

func (rs *DirectGCPStorage) ensureBackupSlotAvailable() error {
	if rs.GCPConfig.MaximumBackupCount == 0 {
		// check is disabled
		return nil
	}
	if rs.client == nil {
		return xerrors.Errorf("no gcloud client available - did you call Init()?")
	}

	bkt := rs.client.Bucket(rs.bucketName())
	ctx := context.Background()
	objs := bkt.Objects(ctx, &gcpstorage.Query{Prefix: fmt.Sprintf("workspaces/%s", rs.WorkspaceName)})

	objcnt := 0
	for {
		_, err := objs.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		objcnt++
	}

	if objcnt > rs.GCPConfig.MaximumBackupCount {
		return xerrors.Errorf("Maximum number of snapshots (%d of %d) reached", objcnt, rs.GCPConfig.MaximumBackupCount)
	}

	return nil
}

func (rs *DirectGCPStorage) trailBackup(ctx context.Context, bkt *gcpstorage.BucketHandle, obj *gcpstorage.ObjectHandle, backupID string, trailLength int) (err error) {
	//nolint:ineffassign
	span, ctx := opentracing.StartSpanFromContext(ctx, "uploadChunk")
	defer tracing.FinishSpan(span, &err)

	trailIter := bkt.Objects(ctx, &gcpstorage.Query{Prefix: rs.trailPrefix()})
	trailingObj := bkt.Object(rs.trailingObjectName(backupID, time.Now()))
	_, err = trailingObj.CopierFrom(obj).Run(ctx)
	if err != nil {
		return
	}
	span.LogKV("trailingBackupDone", trailingObj.ObjectName())
	log.WithField("obj", trailingObj.ObjectName()).Debug("trailing backup done")

	var (
		oldTrailObj *gcpstorage.ObjectAttrs
		trail       []string
	)
	for oldTrailObj, err = trailIter.Next(); oldTrailObj != nil; oldTrailObj, err = trailIter.Next() {
		trail = append(trail, oldTrailObj.Name)
	}
	if err != iterator.Done && err != nil {
		return
	}
	log.WithField("trailLength", len(trail)).Debug("listed backup trail")
	span.LogKV("trailLength", len(trail), "event", "listed backup trail")

	sort.Slice(trail, func(i, j int) bool { return trail[i] < trail[j] })

	for i, oldTrailObj := range trail {
		if i >= len(trail)-trailLength {
			break
		}

		err := bkt.Object(oldTrailObj).Delete(ctx)
		if err != nil {
			span.LogKV("event", "cannot delete trailing backup", "bkt", rs.bucketName(), "obj", oldTrailObj)
			log.WithError(err).WithField("obj", oldTrailObj).Warn("cannot delete old trailing backup")
			continue
		}
		span.LogKV("event", "old trailing object deleted", "bkt", rs.bucketName(), "obj", oldTrailObj)
		log.WithField("obj", oldTrailObj).WithField("originalTrailLength", len(trail)).Debug("old trailing object deleted")
	}
	return nil
}

func (rs *DirectGCPStorage) bucketName() string {
	return gcpBucketName(rs.Stage, rs.Username)
}

// Bucket provides the bucket name for a particular user
func (rs *DirectGCPStorage) Bucket(ownerID string) string {
	return gcpBucketName(rs.Stage, ownerID)
}

// BackupObject returns a backup's object name that a direct downloader would download
func (rs *DirectGCPStorage) BackupObject(name string) string {
	return rs.objectName(name)
}

func gcpBucketName(stage config.Stage, ownerID string) string {
	return fmt.Sprintf("gitpod-%s-user-%s", stage, ownerID)
}

func gcpWorkspaceBackupObjectName(workspaceID string, name string) string {
	return fmt.Sprintf("%s/%s", workspaceID, name)
}

func (rs *DirectGCPStorage) workspacePrefix() string {
	return fmt.Sprintf("workspaces/%s", rs.WorkspaceName)
}

func (rs *DirectGCPStorage) objectName(name string) string {
	return gcpWorkspaceBackupObjectName(rs.workspacePrefix(), name)
}

func (rs *DirectGCPStorage) trailPrefix() string {
	return fmt.Sprintf("%s/trail-", rs.workspacePrefix())
}

func (rs *DirectGCPStorage) trailingObjectName(id string, t time.Time) string {
	return fmt.Sprintf("%s%d-%s", rs.trailPrefix(), t.Unix(), id)
}

func newGCPClient(ctx context.Context, cfg config.GCPConfig) (*gcpstorage.Client, error) {
	credfile := cfg.CredentialsFile
	if tproot := os.Getenv("TELEPRESENCE_ROOT"); tproot != "" {
		credfile = filepath.Join(tproot, credfile)
	}

	client, err := gcpstorage.NewClient(ctx, option.WithCredentialsFile(credfile))
	if err != nil {
		return nil, xerrors.Errorf("cannot create GCP storage client: %w", err)
	}
	return client, nil
}

func newPresignedGCPAccess(config config.GCPConfig, stage config.Stage) (*PresignedGCPStorage, error) {
	err := ValidateGCPConfig(&config)
	if err != nil {
		return nil, xerrors.Errorf("invalid config: %w", err)
	}

	credfile := config.CredentialsFile
	if tproot := os.Getenv("TELEPRESENCE_ROOT"); tproot != "" {
		credfile = filepath.Join(tproot, credfile)
	}

	jsonKey, err := os.ReadFile(credfile)
	if err != nil {
		return nil, xerrors.Errorf("cannot read private key: %w", err)
	}
	privateKey, err := google.JWTConfigFromJSON(jsonKey)
	if err != nil {
		return nil, xerrors.Errorf("cannot get private key: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// We create a client here just to make sure that we can
	client, err := gcpstorage.NewClient(ctx, option.WithCredentialsFile(credfile))
	if err != nil {
		return nil, xerrors.Errorf("cannot create GCP storage client: %w", err)
	}
	client.Close()

	if err != nil {
		return nil, xerrors.Errorf("cannot get Google access ID: %w", err)
	}

	return &PresignedGCPStorage{
		config:     config,
		stage:      stage,
		privateKey: privateKey.PrivateKey,
		accessID:   privateKey.Email,
	}, nil
}

// PresignedGCPStorage provides presigned URLs to access GCP storage objects
type PresignedGCPStorage struct {
	config     config.GCPConfig
	stage      config.Stage
	privateKey []byte
	accessID   string
}

// Bucket provides the bucket name for a particular user
func (p *PresignedGCPStorage) Bucket(owner string) string {
	return gcpBucketName(p.stage, owner)
}

// BlobObject returns a blob's object name
func (p *PresignedGCPStorage) BlobObject(name string) (string, error) {
	return blobObjectName(name)
}

// EnsureExists makes sure that the remote storage location exists and can be up- or downloaded from
func (p *PresignedGCPStorage) EnsureExists(ctx context.Context, bucket string) (err error) {
	client, err := newGCPClient(ctx, p.config)
	if err != nil {
		return err
	}
	//nolint:staticcheck
	defer client.Close()

	return gcpEnsureExists(ctx, client, bucket, p.config)
}

// DiskUsage gives the total objects size of objects that have the given prefix
func (p *PresignedGCPStorage) DiskUsage(ctx context.Context, bucket string, prefix string) (size int64, err error) {
	client, err := newGCPClient(ctx, p.config)
	if err != nil {
		return
	}
	//nolint:staticcheck
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}

	var total int64
	it := client.Bucket(bucket).Objects(ctx, &gcpstorage.Query{
		Prefix: prefix,
	})
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return 0, err
		}
		total += attrs.Size
	}

	return total, nil
}

// SignDownload provides presigned URLs to access remote storage objects
func (p *PresignedGCPStorage) SignDownload(ctx context.Context, bucket, object string, options *SignedURLOptions) (*DownloadInfo, error) {
	client, err := newGCPClient(ctx, p.config)
	if err != nil {
		return nil, err
	}
	//nolint:staticcheck
	defer client.Close()

	bkt := client.Bucket(bucket)
	_, err = bkt.Attrs(ctx)
	if errors.Is(err, gcpstorage.ErrBucketNotExist) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	obj := bkt.Object(object)
	attrs, err := obj.Attrs(ctx)
	if errors.Is(err, gcpstorage.ErrObjectNotExist) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	res, err := p.downloadInfo(ctx, client, attrs, options)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (p *PresignedGCPStorage) downloadInfo(ctx context.Context, client *gcpstorage.Client, obj *gcpstorage.ObjectAttrs, options *SignedURLOptions) (*DownloadInfo, error) {
	meta := &ObjectMeta{
		ContentType:        obj.ContentType,
		OCIMediaType:       obj.Metadata[ObjectAnnotationOCIContentType],
		Digest:             obj.Metadata[ObjectAnnotationDigest],
		UncompressedDigest: obj.Metadata[ObjectAnnotationUncompressedDigest],
	}
	url, err := gcpstorage.SignedURL(obj.Bucket, obj.Name, &gcpstorage.SignedURLOptions{
		Method:         "GET",
		GoogleAccessID: p.accessID,
		PrivateKey:     p.privateKey,
		Expires:        time.Now().Add(30 * time.Minute),
		ContentType:    options.ContentType,
	})
	if err != nil {
		return nil, err
	}

	return &DownloadInfo{
		Meta: *meta,
		URL:  url,
		Size: obj.Size,
	}, nil
}

// SignUpload describes an object for upload
func (p *PresignedGCPStorage) SignUpload(ctx context.Context, bucket, object string, options *SignedURLOptions) (info *UploadInfo, err error) {
	client, err := newGCPClient(ctx, p.config)
	if err != nil {
		return nil, err
	}
	//nolint:staticcheck
	defer client.Close()

	bkt := client.Bucket(bucket)
	_, err = bkt.Attrs(ctx)
	if errors.Is(err, gcpstorage.ErrBucketNotExist) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	url, err := gcpstorage.SignedURL(bucket, object, &gcpstorage.SignedURLOptions{
		Method:         "PUT",
		GoogleAccessID: p.accessID,
		PrivateKey:     p.privateKey,
		Expires:        time.Now().Add(30 * time.Minute),
		ContentType:    options.ContentType,
	})
	if err != nil {
		return nil, err
	}

	return &UploadInfo{
		URL: url,
	}, nil
}

// DeleteObject deletes objects in the given bucket specified by the given query
func (p *PresignedGCPStorage) DeleteObject(ctx context.Context, bucket string, query *DeleteObjectQuery) (err error) {
	client, err := newGCPClient(ctx, p.config)
	if err != nil {
		return err
	}
	//nolint:staticcheck
	defer client.Close()

	if query.Name != "" {
		err = client.Bucket(bucket).Object(query.Name).Delete(ctx)
		if err != nil {
			if errors.Is(err, gcpstorage.ErrBucketNotExist) || errors.Is(err, gcpstorage.ErrObjectNotExist) {
				return ErrNotFound
			}

			log.WithField("bucket", bucket).WithField("object", query.Name).WithError(err).Warn("cannot delete object")
			return err
		}
		return nil
	}

	prefix := query.Prefix
	b := client.Bucket(bucket)
	var it *gcpstorage.ObjectIterator
	if prefix != "" && prefix != "/" {
		it = b.Objects(ctx, &gcpstorage.Query{
			Prefix: prefix,
		})
	} else {
		it = b.Objects(ctx, nil)
	}
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		// if we get any error besides "done" the iterator is broken: make sure we don't use it again.
		if err != nil {
			if errors.Is(err, gcpstorage.ErrBucketNotExist) {
				return ErrNotFound
			}
			log.WithField("bucket", bucket).WithError(err).Error("error iterating object")
			break
		}
		err = b.Object(attrs.Name).Delete(ctx)
		if err != nil {
			if errors.Is(err, gcpstorage.ErrBucketNotExist) || errors.Is(err, gcpstorage.ErrObjectNotExist) {
				continue
			}
			log.WithField("bucket", bucket).WithField("object", attrs.Name).WithError(err).Warn("cannot delete object, continue deleting objects")
		}
	}
	return err
}

// DeleteBucket deletes a bucket
func (p *PresignedGCPStorage) DeleteBucket(ctx context.Context, bucket string) (err error) {
	client, err := newGCPClient(ctx, p.config)
	if err != nil {
		return err
	}
	//nolint:staticcheck
	defer client.Close()

	err = p.DeleteObject(ctx, bucket, &DeleteObjectQuery{})
	if err != nil {
		return err
	}

	err = client.Bucket(bucket).Delete(ctx)
	if err != nil {
		if e, ok := err.(*googleapi.Error); ok {
			if e.Code == http.StatusNotFound {
				return ErrNotFound
			}
		}
		if errors.Is(err, gcpstorage.ErrBucketNotExist) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

// ObjectHash gets a hash value of an object
func (p *PresignedGCPStorage) ObjectHash(ctx context.Context, bucket string, obj string) (hash string, err error) {
	client, err := newGCPClient(ctx, p.config)
	if err != nil {
		return "", err
	}
	//nolint:staticcheck
	defer client.Close()

	attr, err := client.Bucket(bucket).Object(obj).Attrs(ctx)
	if err != nil {
		if errors.Is(err, gcpstorage.ErrBucketNotExist) {
			return "", ErrNotFound
		}
		return "", err
	}
	return hex.EncodeToString(attr.MD5), nil
}

func (p *PresignedGCPStorage) ObjectExists(ctx context.Context, bucket, obj string) (bool, error) {
	client, err := newGCPClient(ctx, p.config)
	if err != nil {
		return false, err
	}
	//nolint:staticcheck
	defer client.Close()

	_, err = client.Bucket(bucket).Object(obj).Attrs(ctx)
	if err != nil {
		if errors.Is(err, gcpstorage.ErrBucketNotExist) {
			return false, nil
		}
		if errors.Is(err, gcpstorage.ErrObjectNotExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// BackupObject returns a backup's object name that a direct downloader would download
func (p *PresignedGCPStorage) BackupObject(ownerID string, workspaceID string, name string) string {
	return fmt.Sprintf("workspaces/%s", gcpWorkspaceBackupObjectName(workspaceID, name))
}

// InstanceObject returns a instance's object name that a direct downloader would download
func (p *PresignedGCPStorage) InstanceObject(ownerID string, workspaceID string, instanceID string, name string) string {
	return p.BackupObject(ownerID, workspaceID, InstanceObjectName(instanceID, name))
}
