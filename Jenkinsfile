node {
    cleanWs()
    def architectures = [
            [pkg: 'jfrog-cli-windows-amd64', goos: 'windows', goarch: 'amd64', fileExtention: '.exe'],
            [pkg: 'jfrog-cli-linux-386', goos: 'linux', goarch: '386', fileExtention: ''],
            [pkg: 'jfrog-cli-linux-amd64', goos: 'linux', goarch: 'amd64', fileExtention: ''],
            [pkg: 'jfrog-cli-linux-arm', goos: 'linux', goarch: 'arm', fileExtention: ''],
            [pkg: 'jfrog-cli-linux-arm64', goos: 'linux', goarch: 'arm64', fileExtention: ''],
            [pkg: 'jfrog-cli-mac-386', goos: 'darwin', goarch: 'amd64', fileExtention: '']
    ]

    subject = 'jfrog'
    repo = 'jfrog-cli'
    sh 'rm -rf temp'
    sh 'mkdir temp'
    def goRoot = tool 'go-1.14'
    env.GOROOT="$goRoot"
    env.PATH+=":${goRoot}/bin"
    env.GO111MODULE="on"
    env.JFROG_CLI_OFFER_CONFIG="false"

    dir('temp') {
        cliWorkspace = pwd()
        sh "echo cliWorkspace=$cliWorkspace"
        stage('Clone JFrog CLI sources') {
            sh 'git clone https://github.com/jfrog/jfrog-cli.git'
            dir("$repo") {
                if (BRANCH?.trim()) {
                    sh "git checkout $BRANCH"
                }
            }
        }

        stage('Build JFrog CLI') {
            jfrogCliRepoDir = "${cliWorkspace}/${repo}/"
            jfrogCliDir = "${jfrogCliRepoDir}jfrog-cli/jfrog"
            sh "echo jfrogCliDir=$jfrogCliDir"

            sh 'go version'
            dir("$jfrogCliRepoDir") {
                sh './build.sh'
            }

            sh 'mkdir builder'
            sh "mv $jfrogCliRepoDir/jfrog builder/"

            // Extract CLI version
            sh 'builder/jfrog --version > version'
            version = readFile('version').trim().split(" ")[2]
            print "CLI version: $version"
        }

        stage('Download tools cert') {
            // Download the certificate file and key file, used for signing the JFrog CLI binary.
            sh """#!/bin/bash
               builder/jfrog rt dl installation-files/certificates/jfrog/ --url https://entplus.jfrog.io/artifactory --flat --access-token=$DOWNLOAD_SIGNING_CERT_ACCESS_TOKEN
                """

            sh 'tar xvzf jfrogltd_signingcer_full.tar.gz'
        }

        stage('Download encryption file') {
            sh """#!/bin/bash
               builder/jfrog rt dl ci-files-local/jfrog-cli/jfrog-cli-enc.json --url https://entplus.jfrog.io/artifactory --flat --access-token=$DOWNLOAD_SIGNING_CERT_ACCESS_TOKEN
                """
            nowIn = pwd()
            sh "echo nowIn=$nowIn"
            sh "mv ./jfrog-cli-enc.json $jfrogCliRepoDir"
        }

        if ("$EXECUTION_MODE".toString().equals("Publish packages")) {
            stage('Npm Publish') {
                print "publishing npm package"
                publishNpmPackage(jfrogCliRepoDir)
            }

            stage('Build and Publish Docker Image') {
                buildPublishDockerImage(version, jfrogCliRepoDir)
            }
        } else if ("$EXECUTION_MODE".toString().equals("Build CLI")) {
            print "Uploading version $version to Bintray"
            uploadCli(architectures)
        }
    }
}

def uploadCli(architectures) {
    for (int i = 0; i < architectures.size(); i++) {
        def currentBuild = architectures[i]
        stage("Build and upload ${currentBuild.pkg}") {
            buildAndUpload(currentBuild.goos, currentBuild.goarch, currentBuild.pkg, currentBuild.fileExtention)
        }
    }
}

def buildPublishDockerImage(version, jfrogCliRepoDir) {
    dir("$jfrogCliRepoDir") {
        docker.build("jfrog-docker-reg2.bintray.io/jfrog/jfrog-cli-go:$version")
        sh '#!/bin/sh -e\n' + 'echo $KEY | docker login --username=$USER_NAME --password-stdin jfrog-docker-reg2.bintray.io/jfrog'
        sh "docker push jfrog-docker-reg2.bintray.io/jfrog/jfrog-cli-go:$version"
        sh "docker tag jfrog-docker-reg2.bintray.io/jfrog/jfrog-cli-go:$version jfrog-docker-reg2.bintray.io/jfrog/jfrog-cli-go:latest"
        sh "docker push jfrog-docker-reg2.bintray.io/jfrog/jfrog-cli-go:latest"
    }
}

def uploadToBintray(pkg, fileName) {
    sh """#!/bin/bash
           builder/jfrog bt u $jfrogCliRepoDir/$fileName $subject/jfrog-cli-go/$pkg/$version /$version/$pkg/ --user=$USER_NAME --key=$KEY
        """
}

def buildAndUpload(goos, goarch, pkg, fileExtension) {
    def extension = fileExtension == null ? '' : fileExtension
    def fileName = "jfrog$fileExtension"
    dir("${jfrogCliRepoDir}") {
        env.GOOS="$goos"
        env.GOARCH="$goarch"
        sh "./build.sh $fileName ./jfrog-cli-enc.json"

        if (goos == 'windows') {
            dir("${cliWorkspace}/certs-dir") {
                // Move the jfrog executable into the 'sign' directory, so that it is signed there.
                sh "mv $jfrogCliRepoDir/$fileName ${jfrogCliRepoDir}sign/${fileName}.unsigned"
                // Copy all the certificate files into the 'sign' directory.
                sh "cp * ${jfrogCliRepoDir}sign/"
                // Build and run the docker container, which signs the JFrog CLI binary.
                sh "docker build -t jfrog-cli-sign-tool ${jfrogCliRepoDir}sign/"
                def signCmd = "osslsigncode sign -certs workspace/JFrog_Ltd_.crt -key workspace/jfrogltd.key  -n JFrog_CLI -i https://www.jfrog.com/confluence/display/CLI/JFrog+CLI -in workspace/${fileName}.unsigned -out workspace/$fileName"
                sh "docker run -v ${jfrogCliRepoDir}sign/:/workspace --rm jfrog-cli-sign-tool $signCmd"
                // Move the JFrog CLI binary from the 'sign' directory, back to its original location.
                sh "mv ${jfrogCliRepoDir}sign/$fileName $jfrogCliRepoDir"
            }
        }

        if (goos == 'linux' && goatch == '386') {
            sh "./$fileName diagnostics"
        }
    }

    uploadToBintray(pkg, fileName)
    sh "rm $jfrogCliRepoDir/$fileName"
}

def publishNpmPackage(jfrogCliRepoDir) {
    dir(jfrogCliRepoDir+'npm/') {
        sh '''#!/bin/bash
            echo "Downloading npm..."
            wget https://nodejs.org/dist/v8.11.1/node-v8.11.1-linux-x64.tar.xz
            tar -xvf node-v8.11.1-linux-x64.tar.xz
            export PATH=$PATH:$PWD/node-v8.11.1-linux-x64/bin/
            echo "//registry.npmjs.org/:_authToken=$NPM_AUTH_TOKEN" > .npmrc
            echo "registry=https://registry.npmjs.org" >> .npmrc
            ./node-v8.11.1-linux-x64/bin/npm publish
        '''
    }
}
