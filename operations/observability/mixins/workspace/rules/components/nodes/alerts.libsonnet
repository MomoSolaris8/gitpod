/**
 * Copyright (c) 2021 Gitpod GmbH. All rights reserved.
 * Licensed under the MIT License. See License-MIT.txt in the project root for license information.
 */

{
  prometheusAlerts+:: {
    groups+: [
      {
        name: 'gitpod-workspace-component-node-alerts',
        rules: [
          {
            alert: 'GitpodWorkspaceNodeHighNormalizedLoadAverage',
            labels: {
              severity: 'warning',
            },
            'for': '2m',
            annotations: {
              runbook_url: 'https://github.com/gitpod-io/observability/blob/main/runbooks/GitpodWorkspaceNodeHighNormalizedLoadAverage.md',
              summary: "Workspace node's normalized load average is higher than 3 for more than 2 minutes. Check for abuse.",
              description: 'Node {{ $labels.node }} is reporting {{ printf "%.2f" $value }}% normalized load average. Normalized load average is current load average divided by number of CPU cores of the node.',
            },
            expr: 'nodepool:node_load1:normalized{nodepool=~".*workspace.*"} > 3',
          },
          {
            alert: 'AutoscalerAddsNodesTooFast',
            labels: {
              severity: 'critical',
            },
            annotations: {
              runbook_url: 'https://github.com/gitpod-io/runbooks/blob/main/runbooks/AutoscalerAddsNodesTooFast.md',
              summary: "Autoscaler is adding new nodes rapidly",
              description: 'Autoscaler in cluster {{ $labels.cluster }} is rapidly adding new nodes.',
            },
            expr: '((sum(cluster_autoscaler_nodes_count) by (cluster)) - (sum(cluster_autoscaler_nodes_count offset 10m) by (cluster))) > 15',
          },
        ],
      },
    ],
  },
}
