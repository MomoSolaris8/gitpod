// Copyright (c) 2022 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package io.gitpod.jetbrains.remote

import com.intellij.openapi.components.service
import com.intellij.openapi.diagnostic.thisLogger
import com.jetbrains.rd.platform.codeWithMe.unattendedHost.metrics.Metric
import com.jetbrains.rd.platform.codeWithMe.unattendedHost.metrics.MetricType
import com.jetbrains.rd.platform.codeWithMe.unattendedHost.metrics.MetricsStatus
import com.jetbrains.rd.platform.codeWithMe.unattendedHost.metrics.providers.MetricProvider
import io.gitpod.jetbrains.remote.utils.Retrier
import io.gitpod.supervisor.api.InfoServiceGrpc
import io.gitpod.supervisor.api.Status.ResourcesStatuRequest
import io.gitpod.supervisor.api.Status.ResourcesStatusResponse
import io.gitpod.supervisor.api.StatusServiceGrpc
import kotlinx.coroutines.GlobalScope
import kotlinx.coroutines.delay
import kotlinx.coroutines.guava.asDeferred
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch
import java.util.concurrent.CancellationException

class GitpodMetricProvider(override val id: String = "gitpodMetricProvider") : MetricProvider {

    private val manager = service<GitpodManager>()

    override fun getMetrics(): Map<String, Metric> {
        val status = manager.status

        /*
        double var4 = (double)var3.getTotalSpace() / (double)1073741824L;
      double var6 = (double)var3.getFreeSpace() / (double)1073741824L;
      double var8 = (double)100 * (var4 - var6) / var4;
      Metric var10 = this.diskMetric(MetricsStatus.NORMAL, var4);
      Metric var11 = this.diskMetric(MetricsStatus.NORMAL, var4 - var6);
      <undefinedtype> var12 = this.diskPercentageMetric(MetricProvider.Companion.calculateMetricStatus((Comparable)var8, (Comparable)80.0, (Comparable)95.0), var8);
         */

        /*
        cpuFraction := int64((float64(workspaceResources.Cpu.Used) / float64(workspaceResources.Cpu.Limit)) * 100)
	    memFraction := int64((float64(workspaceResources.Memory.Used) / float64(workspaceResources.Memory.Limit)) * 100)
	        cpu := fmt.Sprintf("%dm/%dm (%d%%)", workspaceResources.Cpu.Used, workspaceResources.Cpu.Limit, cpuFraction)
	        memory := fmt.Sprintf("%dMi/%dMi (%d%%)\n", workspaceResources.Memory.Used/(1024*1024), workspaceResources.Memory.Limit/(1024*1024), memFraction)
         */

        var cpuValue = 0.0
        var memoryValue = 0.0
        if (status != null) {
            cpuValue = status.cpu.used / status.cpu.limit.toDouble() * 100
            memoryValue = status.memory.used / status.memory.limit.toDouble() * 100
        }
        val cpuStatus = MetricProvider.calculateMetricStatus(cpuValue, 80.0, 95.0)
        val memoryStatus = MetricProvider.calculateMetricStatus(memoryValue, 80.0, 95.0)
        return mapOf(
            "gitpodCpu" to Metric(MetricType.PERFORMANCE, cpuStatus, cpuValue),
            "gitpodMemory" to Metric(MetricType.PERFORMANCE, memoryStatus, memoryValue)
        )
    }
}
