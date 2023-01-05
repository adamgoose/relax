@Library('jenkins-lib') _
pipeline {
    agent any

    options {
        buildDiscarder(logRotator(numToKeepStr: '10', daysToKeepStr: '7'))
        disableConcurrentBuilds()
    }

    triggers {
        githubPush()
    }

    environment
    {
        VERSION = 'latest'
        PROJECT = 'reflektive/chatbot'
        IMAGE = '$PROJECT:latest'
        ECRURL = 'https://965162577299.dkr.ecr.us-east-1.amazonaws.com'
        ECRCRED = 'ecr:us-east-1:2a3c567b-dcfb-4cef-a18f-9d885a22edfd'
    }

    stages {

        stage('Initialize') {
            steps {
                echo "Clean WS"
                deleteDir()
            }
        }

        stage ('Checkout and Prepare') {
            steps {
              dir('platform-devops') {
                git branch: 'reflektive_main',
                    credentialsId: '97bccddc-4590-4b47-93c9-3ce9a306aa2c',
                    poll: false,
                    url: 'git@github.com:PeopleFluent/platform_devops.git'
              }
              dir('relax') {
                // Checkout the branch
                git credentialsId: '7c77061f-a395-48c7-84d1-12f33c68eb41',
                    branch: "${GIT_BRANCH}",
                    url: 'git@github.com:PulseSoftwareInc/relax.git'

                // checkout scm

                script {
                    //Get git commit short-hash and set version
                    SHORT_HASH     = sh(returnStdout: true, script: 'git rev-parse --short HEAD').trim()
                    CUSTOM_DATE    = sh(returnStdout: true, script: 'echo $(TZ=Asia/Kolkata date "+%m-%d-%H%M")').trim()
                    VERSION        = "$SHORT_HASH-$CUSTOM_DATE-$BRANCH_NAME"
                    IMAGE          = "$PROJECT:$VERSION"
                    echo "$SHORT_HASH"
                    sh 'tail ../platform-devops/k8s/reflektive/service-deployments/chatbot/values-chatbot.yaml'
                    EXISTING_IMAGE_COUNT = sh(returnStdout: true, script: "grep $SHORT_HASH.*.$BRANCH_NAME ../platform-devops/k8s/reflektive/service-deployments/chatbot/values-chatbot.yaml | wc -l" ).trim()
                    echo "$EXISTING_IMAGE_COUNT"
                }
              }
            }
        }

        stage('Verify') {
            steps {
                script {
                    def check = "${EXISTING_IMAGE_COUNT}" as int
                    echo "checking image count: $check"
                    if (check < 1) {
                        echo "$EXISTING_IMAGE_COUNT images found, building fresh : "
                        stage ('Build Image') {
                            dir('relax') {
                                script {
                                    docker.withRegistry("$ECRURL", "$ECRCRED") {
                                        docker.build("$IMAGE", "--no-cache .")
                                    }
                                }
                            }
                        }

                        stage ('Push Image') {
                            dir('relax') {
                                script {
                                    docker.withRegistry("$ECRURL", "$ECRCRED") {
                                        docker.image("$IMAGE").push()
                                    }
                                }
                            }
                        }

                        stage ('Delete Image') {
                            script {
                                sh "docker container prune -f"
                                sh "docker rmi -f \$(docker images | grep \${PROJECT} | awk '{print \$3}') || :"
                            }
                        }
                    }
                    else {
                        echo "$EXISTING_IMAGE_COUNT images found, skipping all the stages..."
                    }
                }
            }
        }


        stage('Deployment for Integration') {
            when {
                branch 'dev'
            }
            steps {
                dir('platform-devops') {
                    script {
                        def check = "${EXISTING_IMAGE_COUNT}" as int
                        echo "checking image count: $check"
                        if (check < 1) {
                            echo "$EXISTING_IMAGE_COUNT images found, deploying..."
                            sh """
                                mkdir -p /tmp/relax
                                cp k8s/reflektive/service-deployments/chatbot/values-chatbot.yaml /tmp/relax/
                                sed -i.bak 's#^devtag:.*#devtag: $VERSION#' /tmp/relax/values-chatbot.yaml
                                git config user.email 'parthiban.rengaraj@peoplefluent.com'
                                git config user.name 'parthiops'
                                cp /tmp/relax/values-chatbot.yaml k8s/reflektive/service-deployments/chatbot/
                                git commit -a -m 'Build-${VERSION}'
                                """

                            withCredentials([sshUserPrivateKey(credentialsId: '97bccddc-4590-4b47-93c9-3ce9a306aa2c', keyFileVariable: 'SSH_KEY')]) {
                                sh('GIT_SSH_COMMAND="ssh -i $SSH_KEY" git push origin reflektive_main')
                            }
                        }
                        else {
                            echo "$EXISTING_IMAGE_COUNT images found, skipping all the stages..."
                        }
                    }
                }
            }
        }

        // stage('Deployment for Staging-Demo-Preview') {
        //     when { anyOf { branch 'release';  branch 'temp_r'} }
        //     steps {
        //         dir('platform-devops') {
        //             script {
        //                 def check = "${EXISTING_IMAGE_COUNT}" as int
        //                 echo "checking image count: $check"
        //                 if (check < 1) {
        //                     echo "$EXISTING_IMAGE_COUNT images found, deploying..."
        //                     sh """
        //                         mkdir -p /tmp/pulse360
        //                         cp k8s/reflektive/service-deployments/pulse/values-pulse.yaml /tmp/pulse360/
        //                         sed -i.bak 's#^stagingtag:.*#stagingtag: $VERSION#' /tmp/pulse360/values-pulse.yaml
        //                         sed -i.bak 's#^demotag:.*#demotag: $VERSION#' /tmp/pulse360/values-pulse.yaml
        //                         sed -i.bak 's#^previewtag:.*#previewtag: $VERSION#' /tmp/pulse360/values-pulse.yaml
        //                         sed -i.bak 's#^gsdevtag:.*#gsdevtag: $VERSION#' /tmp/pulse360/values-pulse.yaml
        //                         sed -i.bak 's#^gsuattag:.*#gsuattag: $VERSION#' /tmp/pulse360/values-pulse.yaml
        //                         git config user.email 'parthiban.rengaraj@peoplefluent.com'
        //                         git config user.name 'parthiops'
        //                         cp /tmp/pulse360/values-pulse.yaml k8s/reflektive/service-deployments/pulse/
        //                         git commit -a -m 'Build-${VERSION}'
        //                         """
        //                     withCredentials([sshUserPrivateKey(credentialsId: '97bccddc-4590-4b47-93c9-3ce9a306aa2c', keyFileVariable: 'SSH_KEY')]) {
        //                             sh('GIT_SSH_COMMAND="ssh -i $SSH_KEY" git push origin reflektive_main')
        //                     }
        //                 }
        //                 else {
        //                     echo "$EXISTING_IMAGE_COUNT images found, skipping all the stages..."
        //                 }
        //             }
        //         }
        //     }
        // }
    }
}
