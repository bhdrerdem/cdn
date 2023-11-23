import * as dotenv from "dotenv";
dotenv.config();

const envOrElse = (name: string, elseVal: string): string => {
    const val = process.env[name];
    return val ?? elseVal;
};

const intEnvOrElse = (name: string, elseVal: number): number => {
    const val = envOrElse(name, elseVal.toString());
    const parsed = parseInt(val, 10);
    return isNaN(parsed) ? elseVal : parsed;
};

export interface CdnAppConfig {
    // VPC
    vpcMaxAzs: number;
    vpcNatGateways: number;

    // ECS
    ecsTaskCpu: number;
    ecsTaskMemory: number;
    ecsContainerImage: string;
    ecsContainerPort: number;
}

export const config: CdnAppConfig = {
    vpcMaxAzs: intEnvOrElse("VPC_MAX_AZS", 2),
    vpcNatGateways: intEnvOrElse("VPC_NAT_GATEWAYS", 1),
    ecsTaskCpu: intEnvOrElse("ECS_TASK_CPU", 256),
    ecsTaskMemory: intEnvOrElse("ECS_TASK_MEMORY", 512),
    ecsContainerImage: envOrElse("IMAGE", ""),
    ecsContainerPort: intEnvOrElse("ECS_CONTAINER_PORT", 8080),
};
