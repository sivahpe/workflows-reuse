# Container build template

Initialize a new repo for SLSA3 compliant container image publishing.

## How to use this template

1. Select the green button labeled `Use this template` and then in the drop-down select `Create a new repository`
1. Input a name for the new repository and select the green button labeled `Create repository from template`
1. In the new repository, replace or edit the Dockerfile with your own content and update the README.md
1. Fix the repo settings for best practices
   1. Unselect `Allow merge commits` [optional]
   1. Unselect `Allow rebase merging` [optional]
   1. Select `Always suggest updating pull request branches`
   1. Select `Allow auto-merge` [optional]
   1. Select `Automatically delete head branches`
   1. Add branch protection rules for the default branch
      1. Select `Require a pull request before merging`
         1. Select `Dismiss stale pull request approvals when new commits are pushed`
         1. Select `Require review from Code Owners` (also add `.github/codeowners` file)
      1. Select `Require status checks to pass before merging`
         1. Add the following status checks:
            1. `secrets-scanner / secret-scanner`
            1. `copyright / copyright` (must create a PR for this check to exist)
            1. `container-build`
      1. Select `Require signed commits` [optional]
      1. Select `Do not allow bypassing the above settings`
    1. Add a team to the repo and grant write permissions so that they can maintain PRs
    1. Add a team to the repo and grant admin permissions

## How to pull an image

1. A developer can be authorized using these [instructions](https://github.com/hpe-sre/image-puller#readme)
1. AWS accounts vended from CCP and GLC organizations are authorized to read from the registries by granting the [AWS managed ECR read-only role](https://docs.aws.amazon.com/AmazonECR/latest/userguide/security-iam-awsmanpol.html#security-iam-awsmanpol-AmazonEC2ContainerRegistryReadOnly) to the principal, e.g. service
1. GitHub actions for organizations in the HPE enterprise can be authorized to read from the registries by including the following snippet:
   ```
   jobs:
     ecr-login:
       runs-on: ubuntu-latest
       permissions:
         id-token: write
         contents: read
       steps:
         - name: Checkout
           uses: actions/checkout@93ea575cb5d8a053eaa0ac8fa3b40d7e05a33cc8
           with:
             fetch-depth: 0
         - name: Configure AWS Credentials
           id: assume-role
           uses: aws-actions/configure-aws-credentials@67fbcbb121271f7775d2e7715933280b06314838
           with:
             role-to-assume: arn:aws:iam::*<your registry aws account number>*:role/github/oidc/ecr-readonly
             role-session-name: github-actions
             aws-region: us-west-2
         - name: Get Amazon ECR Login Token
           id: ecr
           uses: aws-actions/amazon-ecr-login@261a7de32bda11ba01f4d75c4ed6caf3739e54be
           with:
             registries: "*<your registry aws account number>*"
       outputs:
         repository-username: ${{ steps.ecr.outputs.docker_username_*<your aws account number>*_dkr_ecr_us_west_2_amazonaws_com }}
         repository-password: ${{ steps.ecr.outputs.docker_password_*<your aws account number>*_dkr_ecr_us_west_2_amazonaws_com }}

     *my_job*:
       runs-on: ubuntu-latest
       permissions:
         id-token: write
         contents: read
       needs: ecr-login
       container:
         image: *<image to use>*
         credentials:
           username: ${{ needs.ecr-login.outputs.repository-username }}
           password: ${{ needs.ecr-login.outputs.repository-password }}
   ```
