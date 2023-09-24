// Copyright 2022 Hewlett Packard Enterprise Development LP

def getHarborRegistry(){
    ( harborRegistry, unused ) = getHarborPushRegistryAndCred()
    return harborRegistry
}

def getHarborRegistry

pipeline {
    agent {
        kubernetes {
            inheritFrom 'cds-go-dev'
            defaultContainer 'cds-go-dev'
		}
    }
    environment {
        REPO_NAME   = 'pipeline_test'
        PROJECT_NAME= 'hello-world'
        CHART_NAMES = 'pipeline_test'
        PROXY = 'http://web-proxy.corp.hpecorp.net:8080'
        HTTPS_PROXY = "${PROXY}"
        HTTP_PROXY = "${PROXY}"
        NO_PROXY    = '127.0.0.1,localhost,0.0.0.0,.nimblestorage.com,10.0.0.0/8,cloud-objectstore-manager-grpc,ccs-pg,ccs-localstack,prometheus-pushgateway'
        KUBECTL_CONTAINER_VERSION = "1.20-debian-10"
    }

    stages{
        stage('Lint + Trivy scan') {
            parallel {
                stage('Lint') {
                    steps {
                        sh '''
                            cd tools/pipeline_test/
                            make lint
                            '''
                    }
                }

                stage('Trivy FileSystem Scan') {
                    agent {
                        kubernetes {
                            inheritFrom 'trivy'
                            defaultContainer 'trivy'
                        }
                    }
                    steps {
                        withCredentials([
                            string(
                                credentialsId: 'trivy_github_token',
                                variable: 'GITHUB_TOKEN'
                            )
                        ])
                        {
                            sh 'trivy --version'
                            sh 'trivy fs --exit-code 0 --severity UNKNOWN,LOW,MEDIUM --no-progress tools/pipeline_test/'
                            sh 'trivy fs --exit-code 1 --severity HIGH,CRITICAL --no-progress tools/pipeline_test/'
                        }
                    }
                }
            }
        }

        stage('Kube-score') {
            agent {
                kubernetes {
                    inheritFrom 'cds-backend-base'
                    defaultContainer 'cds-backend-base'
                    yaml '''
apiVersion: "v1"
kind: "Pod"
spec:
  containers:
  - name: "kube-score"
    image: "vision-harbor.rtplab.nimblestorage.com/docker_proxy/zegl/kube-score:latest-helm3"
    imagePullPolicy: "Always"
    command:
    - "sleep"
    args:
    - "9999999"
    resources:
      limits: {}
      requests: {}
    tty: false
    volumeMounts:
    - mountPath: "/home/jenkins/agent"
      name: "workspace-volume"
      readOnly: false
      workingDir: "/home/jenkins/agent"
'''
                }
            }

            steps {
                container('cds-backend-base') {
                    sh '''
                        cd tools/pipeline_test
                        helm template helm/hello-world >> rendered.yaml
                        cat rendered.yaml
                    '''
                }
                container('kube-score') {
                    sh '''
                        cd tools/pipeline_test
                        kube-score score rendered.yaml --ignore-test pod-networkpolicy,networkpolicy-targets-pod,container-ephemeral-storage-request-and-limit
                    '''
                }
            }
        }

        stage('Tag') {
            steps {
                script {
                    if (env.GIT_BRANCH == 'master' && !env.CHANGE_TARGET) {
                        withCredentials([
                            usernamePassword(
                                credentialsId: 'sc-github-app',
                                usernameVariable: 'GITHUB_APP',
                                passwordVariable: 'TAG_MANAGE_TOKEN',
                            )
                        ]) {
                            sh '''
                                git config user.name cds-github-ci
                                git config user.email cds-github-ci@hpe.com
                                tag-manage create --min 0.1.0 --push -vv
                            '''
                        }
                    }

                    def version = sh(script: 'tag-manage describe --default 0.0.0', returnStdout: true).trim()
                    env.VERSION = version
                    currentBuild.description = "Version: ${version}"
                }
            }
        }

        stage('SonarQube') {
            agent {
                kubernetes {
                    inheritFrom 'sonarqube-scanner'
                    defaultContainer 'sonarqube-scanner'
                }
            }
            environment{
                https_proxy = "${PROXY}"
                http_proxy = "${PROXY}"
                no_proxy  = "${NO_PROXY}"
            }
            steps {
                withCredentials([
                    string(
                        credentialsId: 'vision-sonarqube-token',
                        variable: 'SECRET'
                        )
                    ]) {
                        sh """
                        cd tools/pipeline_test
                        sonar-scanner \
                        -Dsonar.login=${env.SECRET} \
                        -Dsonar.projectKey=${env.REPO_NAME} \
                        -Dsonar.projectVersion=${env.VERSION} \
                        -Dsonar.host.url=${env.SONARQUBE_URL} \
                        -Dsonar.sources=. 
                        """
                    }
            }
        }

        stage('Build Docker images') {
            agent {
                kubernetes {
                    cloud 'vision'
                    inheritFrom 'k8s-dind'
                }
            }
            steps {
                container('docker-daemon') {
                    script {
                        def ( harborPushRegistry, unused ) = getHarborPushRegistryAndCred()
                        sh '''
                            cd tools/pipeline_test
                            make docker-build
                        '''

                        docker.withRegistry(env.CDS_HARBOR_REGISTRY, unused) {
                            sh """
                            cd tools/pipeline_test
                            make docker-push \
                                PUSH_REGISTRY=${env.HARBOR_HOST} \
                                PUSH_PROJECT=${harborPushRegistry} \
                                PUSH_TAG=${env.VERSION}
                            """
                        }
                    }
                }
            }
        }

        stage('Build arm64 and amd64 Docker images') {
            agent {
                kubernetes {
                    cloud 'vision'
                    inheritFrom 'k8s-dind'
                }
            }
            steps {
                container('docker-daemon') {
                    script {
                        def ( harborPushRegistry, unused ) = getHarborPushRegistryAndCred()
                        docker.withRegistry(env.CDS_HARBOR_REGISTRY, unused) {
                            setupDockerBuildxEnv()
                            sh """
                                cd tools/pipeline_test
                                make docker-build-multiple-targets \
                                    PUSH_REGISTRY=${env.HARBOR_HOST} \
                                    PUSH_PROJECT=${harborPushRegistry} \
                                    PUSH_TAG=${env.VERSION}
                            """
                        }
                    }
                }
            }
        }

        stage('Sign Docker Image') {
            agent{
                kubernetes {
                    inheritFrom 'cosign'
                }
            }
            steps {
                container('cosign') {
                    script {
                        def ( harborPushRegistry, harborPushCred ) = getHarborPushRegistryAndCred()
                        docker.withRegistry(env.VISION_HARBOR_REGISTRY, harborPushCred) {
                            withCredentials([string(credentialsId: 'cosign_password', variable: 'COSIGN_PASSWORD')]) {
                                withCredentials([file(credentialsId: 'cosign-private-key', variable: 'KEY')]) {
                                    def ImageUrl = "${env.HARBOR_HOST}/${harborRegistry}/${env.PROJECT_NAME}:${env.VERSION}"
                                    sh 'cosign sign --key $KEY' + " ${ImageUrl}"
                                }
                            }
                        }
                    }
                }
            }
        }

        stage('Container Scan') {
            agent {
                kubernetes {
                    inheritFrom 'container-scan'
                    defaultContainer 'harborctl'
                }
            }
            steps {
                withCredentials([
                    usernamePassword(
                        credentialsId: 'vision-harbor-robot', 
                        usernameVariable: 'HARBORCTL_USERNAME', 
                        passwordVariable: 'HARBORCTL_PASSWORD'
                    )
                ]) {
                    sh "harborctl scan results vision-harbor.rtplab.nimblestorage.com/sc-jenkins-branch/${env.PROJECT_NAME}:${env.VERSION} --max-critical-count=0 --max-high-count=0"
                }
            }
        }

        stage("Build and publish Helm charts") {
            agent {
                kubernetes {
                    inheritFrom 'sc-helm-s3'
                }
            }
            steps {
                container('sc-helm-s3') {
                    script {
                        dir('tools/pipeline_test/helm') {
                            sh "helm package hello-world --version ${env.VERSION} --app-version ${env.VERSION}"

                            if (env.GIT_BRANCH == 'master') {
                                sh "helm push --force ${env.PROJECT_NAME}-${env.VERSION}.tgz vision-jenkins-preprod"
                            } else {
                                sh "helm push --force ${env.PROJECT_NAME}-${env.VERSION}.tgz vision-jenkins-branch"
                            }
                        }
                    }
                }
            }
        }
    }

    post {
        unsuccessful {
            script {
                if (env.GIT_BRANCH == 'master') {
                    // NB Update to dscc-infra-alerts-channel
                    // chatNotification(channel: '')
                    sh "echo \'unsuccessful\'"
                }
            }
        }
    }
}
