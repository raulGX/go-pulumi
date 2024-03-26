import * as pulumi from "@pulumi/pulumi";
import * as awsx from "@pulumi/awsx";
import * as aws from "@pulumi/aws";
import * as eks from "@pulumi/eks";
import * as k8s from "@pulumi/kubernetes";

const config = new pulumi.Config();
const minClusterSize = config.getNumber("minClusterSize") || 3;
const maxClusterSize = config.getNumber("maxClusterSize") || 6;
const desiredClusterSize = config.getNumber("desiredClusterSize") || 3;
const eksNodeInstanceType = config.get("eksNodeInstanceType") || "t3.micro";
const vpcNetworkCidr = config.get("vpcNetworkCidr") || "10.0.0.0/16";

const eksVpc = new awsx.ec2.Vpc("eks-vpc", {
  enableDnsHostnames: true,
  cidrBlock: vpcNetworkCidr,
});

// Create the EKS cluster
const cluster = new eks.Cluster("cluster", {
  vpcId: eksVpc.vpcId,
  version: "1.29",
  publicSubnetIds: eksVpc.publicSubnetIds,
  privateSubnetIds: eksVpc.privateSubnetIds,
  instanceType: eksNodeInstanceType,
  desiredCapacity: desiredClusterSize,
  minSize: minClusterSize,
  maxSize: maxClusterSize,
  nodeAssociatePublicIpAddress: false,
  endpointPrivateAccess: false,
  endpointPublicAccess: true,
  createOidcProvider: true,
});

const { provider } = cluster;
const ingressChart = new k8s.helm.v3.Release(
  "ingress-nginx",
  {
    namespace: "ingress",
    version: "4.10.0",
    createNamespace: true,
    name: "ingress-nginx",
    chart: "ingress-nginx",
    repositoryOpts: {
      repo: "https://kubernetes.github.io/ingress-nginx",
    },
    values: {
      controller: {
        service: {
          annotations: {
            "service.beta.kubernetes.io/aws-load-balancer-type": "nlb",
          },
        },
      },
    },
  },
  { provider }
);
const ciNamespace = "ci";
const namespace = new k8s.core.v1.Namespace(ciNamespace, {
  metadata: {
    name: ciNamespace,
  },
});
const runnerSecretName = "ci-runner";
const runnerSecret = new k8s.core.v1.Secret(
  runnerSecretName,
  {
    type: "opaque",
    metadata: {
      name: runnerSecretName,
      namespace: namespace.metadata.name,
    },
    stringData: {
      github_token: config.requireSecret("githubToken"),
    },
  },
  {
    provider,
    deleteBeforeReplace: true,
  }
);
const runnerLabels = {
  type: "ci-runner",
};

let serviceAccount = new k8s.core.v1.ServiceAccount(
  "runner-sa",
  {
    metadata: {
      namespace: namespace.metadata.name,
      name: "runner-sa",
    },
  },
  {
    provider,
    deleteBeforeReplace: true,
  }
);

let clusterRole = new k8s.rbac.v1.ClusterRole(
  "runner-role",
  {
    rules: [
      {
        apiGroups: ["*"],
        resources: ["*"],
        verbs: ["get", "list", "watch", "create", "update", "patch", "delete"],
      },
    ],
  },
  {
    provider,
    deleteBeforeReplace: true,
  }
);

let clusterRoleBinding = new k8s.rbac.v1.ClusterRoleBinding(
  "pod-editor-binding",
  {
    subjects: [
      {
        kind: "ServiceAccount",
        name: serviceAccount.metadata.name,
        namespace: serviceAccount.metadata.namespace,
      },
    ],
    roleRef: {
      kind: "ClusterRole",
      name: clusterRole.metadata.name,
      apiGroup: "rbac.authorization.k8s.io",
    },
  },
  {
    provider,
    deleteBeforeReplace: true,
  }
);

const certManager = new k8s.helm.v3.Release("cert-manager", {
  chart: "cert-manager",
  version: "1.5.3",
  namespace: namespace.metadata.name,
  repositoryOpts: {
    repo: "https://charts.jetstack.io",
  },
  values: {
    installCRDs: true,
  },
});

const githubRunnerRelease = new k8s.helm.v3.Release(
  "cluster-ci",
  {
    chart: "actions-runner-controller",
    namespace: namespace.metadata.name,
    repositoryOpts: {
      repo: "https://actions-runner-controller.github.io/actions-runner-controller",
    },
    values: {
      githubConfigSecret: runnerSecret.metadata.name,
    },
  },
  {
    provider,
    dependsOn: [certManager, serviceAccount],
  }
);

const repositoryA = new aws.ecr.Repository("service-a");
const repositoryB = new aws.ecr.Repository("service-b");

export const vpcId = eksVpc.vpcId;
