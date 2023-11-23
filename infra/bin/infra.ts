#!/usr/bin/env node
import "source-map-support/register";
import * as cdk from "aws-cdk-lib";
import { CdnAppStack } from "../lib/infra-stack";
import { config } from "../config/default";

const app = new cdk.App();
new CdnAppStack(app, "CdnAppStack", { ...config });
