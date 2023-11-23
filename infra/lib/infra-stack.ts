import { App, RemovalPolicy, Stack, StackProps } from "aws-cdk-lib";
import {
    Vpc,
    SubnetType,
    SecurityGroup,
    Peer,
    Port,
} from "aws-cdk-lib/aws-ec2";
import {
    Cluster,
    ContainerImage,
    FargateTaskDefinition,
    FargateService,
} from "aws-cdk-lib/aws-ecs";
import { Bucket } from "aws-cdk-lib/aws-s3";
import { Distribution } from "aws-cdk-lib/aws-cloudfront";
import { ApplicationLoadBalancer } from "aws-cdk-lib/aws-elasticloadbalancingv2";
import { S3Origin } from "aws-cdk-lib/aws-cloudfront-origins";
import * as ecs from "aws-cdk-lib/aws-ecs";
import {
    ApplicationTargetGroup,
    ApplicationProtocol,
    TargetType,
} from "aws-cdk-lib/aws-elasticloadbalancingv2";
import { Role, ServicePrincipal, ManagedPolicy } from "aws-cdk-lib/aws-iam";
import { LogGroup, RetentionDays } from "aws-cdk-lib/aws-logs";
import { CdnAppConfig } from "../config/default";

export interface CdnAppStackProps extends StackProps, CdnAppConfig {}

export class CdnAppStack extends Stack {
    constructor(scope: App, id: string, props?: CdnAppStackProps) {
        super(scope, id, props);

        // Define a VPC
        const vpc = new Vpc(this, "CdnVpc", {
            maxAzs: props?.vpcMaxAzs,
            natGateways: props?.vpcNatGateways,
            subnetConfiguration: [
                {
                    cidrMask: 24,
                    name: "cdnPublicSubnet",
                    subnetType: SubnetType.PUBLIC,
                },
                {
                    cidrMask: 24,
                    name: "cdnPrivateSubnet",
                    subnetType: SubnetType.PRIVATE_WITH_EGRESS,
                },
            ],
        });

        // Create an ECS Cluster
        const cluster = new Cluster(this, "CdnCluster", { vpc });

        // Define the ECS Task Security Group
        const taskSecurityGroup = new SecurityGroup(
            this,
            "CdnTaskSecurityGroup",
            {
                vpc,
                allowAllOutbound: true,
                description: "Security group for cdn tasks",
            }
        );

        // Create a security group for the ALB
        const albSecurityGroup = new SecurityGroup(
            this,
            "CdnAlbSecurityGroup",
            {
                vpc,
                allowAllOutbound: true,
                description: "Security group for cdn alb",
            }
        );

        albSecurityGroup.addIngressRule(Peer.anyIpv4(), Port.tcp(80));
        albSecurityGroup.addIngressRule(Peer.anyIpv4(), Port.tcp(443));

        // Create an Application Load Balancer
        const alb = new ApplicationLoadBalancer(this, "CdnALB", {
            vpc,
            internetFacing: true,
            securityGroup: albSecurityGroup,
            vpcSubnets: {
                subnets: vpc.publicSubnets,
            },
        });

        // Add a listener to the ALB
        const listener = alb.addListener("Listener", { port: 80 });

        taskSecurityGroup.addIngressRule(
            albSecurityGroup,
            Port.tcp(props!.ecsContainerPort)
        );

        // Create an S3 bucket for storage
        const bucket = new Bucket(this, "CdnBucket");

        // Create a CloudFront distribution for content delivery
        const distribution = new Distribution(this, "CdnDistribution", {
            defaultBehavior: {
                origin: new S3Origin(bucket),
            },
        });

        const taskExecRole = new Role(this, "cdnTaskExecutionRole", {
            assumedBy: new ServicePrincipal("ecs-tasks.amazonaws.com"),
            managedPolicies: [
                ManagedPolicy.fromAwsManagedPolicyName(
                    "service-role/AmazonECSTaskExecutionRolePolicy"
                ),
            ],
        });

        const logGroup = new LogGroup(this, "CdnLogGroup", {
            retention: RetentionDays.ONE_MONTH,
            removalPolicy: RemovalPolicy.DESTROY,
        });

        // Create an ECS Fargate Task Definition
        const taskDefinition = new FargateTaskDefinition(this, "CdnTaskDef", {
            memoryLimitMiB: props!.ecsTaskMemory,
            cpu: props!.ecsTaskCpu,
            runtimePlatform: {
                cpuArchitecture: ecs.CpuArchitecture.ARM64,
                operatingSystemFamily: ecs.OperatingSystemFamily.LINUX,
            },
            executionRole: taskExecRole,
        });

        // Add a container to the Task Definition
        const container = taskDefinition.addContainer("CdnContainer", {
            image: ContainerImage.fromRegistry(props!.ecsContainerImage),
            environment: {
                PORT: props!.ecsContainerPort.toString(),
                BUCKET: bucket.bucketName,
                DISTRIBUTION_ID: distribution.distributionId,
                DISTRIBUTION_URL: distribution.distributionDomainName,
            },
            logging: ecs.LogDriver.awsLogs({
                streamPrefix: "ecs",
                logGroup: logGroup,
            }),
        });

        // Add port mappings to the container
        container.addPortMappings({
            containerPort: props!.ecsContainerPort,
            protocol: ecs.Protocol.TCP,
        });

        // Create an ECS Service
        const service = new FargateService(this, "CdnService", {
            cluster,
            taskDefinition,
            securityGroups: [taskSecurityGroup],
            vpcSubnets: {
                subnets: vpc.privateSubnets,
            },
            assignPublicIp: false,
            desiredCount: 2,
        });

        // Create a target group for the ECS service
        const targetGroup = new ApplicationTargetGroup(this, "CdnTargetGroup", {
            vpc,
            port: props!.ecsContainerPort,
            protocol: ApplicationProtocol.HTTP,
            targetType: TargetType.IP,
            healthCheck: {
                path: "/health",
                port: props!.ecsContainerPort.toString(),
            },
        });

        // Attach the target group to the listener
        listener.addTargetGroups("CdnListenerTargets", {
            targetGroups: [targetGroup],
        });

        service.attachToApplicationTargetGroup(targetGroup);

        // Grant the ECS task access to the S3 bucket
        bucket.grantReadWrite(taskDefinition.taskRole);
    }
}
