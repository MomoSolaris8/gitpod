/**
 * Copyright (c) 2022 Gitpod GmbH. All rights reserved.
 * Licensed under the GNU Affero General Public License (AGPL).
 * See License-AGPL.txt in the project root for license information.
 */

import { useEffect, useState } from "react";
import Modal from "../components/Modal";
import Tooltip from "../components/Tooltip";
import copy from "../images/copy.svg";
import Alert from "../components/Alert";
import { getGitpodService } from "../service/service";

function InputWithCopy(props: { value: string; tip?: string; className?: string }) {
    const [copied, setCopied] = useState<boolean>(false);
    const copyToClipboard = (text: string) => {
        const el = document.createElement("textarea");
        el.value = text;
        document.body.appendChild(el);
        el.select();
        try {
            document.execCommand("copy");
        } finally {
            document.body.removeChild(el);
        }
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };
    const tip = props.tip ?? "Click to copy";
    return (
        <div className={`w-full relative ${props.className ?? ""}`}>
            <input
                disabled={true}
                readOnly={true}
                autoFocus
                className="w-full pr-8 overscroll-none"
                type="text"
                value={props.value}
            />
            <div className="cursor-pointer" onClick={() => copyToClipboard(props.value)}>
                <div className="absolute top-1/3 right-3">
                    <Tooltip content={copied ? "Copied" : tip}>
                        <img src={copy} alt="copy icon" title={tip} />
                    </Tooltip>
                </div>
            </div>
        </div>
    );
}

interface SSHProps {
    workspaceId: string;
    ownerToken: string;
    ideUrl: string;
}

function SSHView(props: SSHProps) {
    const [hasSSHKey, setHasSSHKey] = useState(false);

    useEffect(() => {
        getGitpodService()
            .server.hasSSHPublicKey()
            .then((d) => {
                setHasSSHKey(d);
            })
            .catch(console.error);
    }, []);

    const host = props.ideUrl.replace(props.workspaceId, props.workspaceId + ".ssh");
    const sshPswCommand = `ssh '${props.workspaceId}#${props.ownerToken}@${host}'`;
    const sshKeyCommand = `ssh '${props.workspaceId}@${host}'`;

    return (
        <div>
            <div className="space-y-4">
                <Alert type="warning" className="whitespace-normal">
                    <b>Anyone</b> on the internet with this command can access the running workspace. The command
                    includes a generated access token that resets on every workspace restart.
                </Alert>
                <Alert type="info" className="whitespace-normal">
                    Before connecting via SSH, make sure you have an existing SSH private key on your machine. You can
                    create one using&nbsp;
                    <a
                        href="https://en.wikipedia.org/wiki/Ssh-keygen"
                        target="_blank"
                        rel="noopener noreferrer"
                        className="gp-link"
                    >
                        ssh-keygen
                    </a>
                    .
                </Alert>
                <p className="text-gray-500 whitespace-normal text-base">
                    The following shell command can be used to SSH into this workspace.
                </p>
            </div>
            <InputWithCopy className="my-2" value={hasSSHKey ? sshKeyCommand : sshPswCommand} tip="Copy SSH Command" />
        </div>
    );
}

export default function ConnectToSSHModal(props: {
    workspaceId: string;
    ownerToken: string;
    ideUrl: string;
    onClose: () => void;
}) {
    return (
        <Modal
            title="Connect via SSH"
            buttons={
                <button className={"ml-2 secondary"} onClick={() => props.onClose()}>
                    Close
                </button>
            }
            visible={true}
            onClose={props.onClose}
        >
            <SSHView workspaceId={props.workspaceId} ownerToken={props.ownerToken} ideUrl={props.ideUrl} />
        </Modal>
    );
}
