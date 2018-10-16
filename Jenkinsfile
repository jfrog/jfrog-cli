node {
    cleanWs()
    def architectures = [
            [pkg: 'jfrog-cli-linux-386', goos: 'linux', goarch: '386', fileExtention: ''],
            [pkg: 'jfrog-cli-linux-amd64', goos: 'linux', goarch: 'amd64', fileExtention: ''],
            [pkg: 'jfrog-cli-mac-386', goos: 'darwin', goarch: 'amd64', fileExtention: ''],
            [pkg: 'jfrog-cli-windows-amd64', goos: 'windows', goarch: 'amd64', fileExtention: '.exe']
    ]

    subject = 'jfrog'
    repo = 'jfrog-cli-go'
    sh 'rm -rf temp'
    sh 'mkdir temp'
    def goRoot = tool 'go-1.11'
    dir('temp') {
        cliWorkspace = pwd()
        jfrogCliParentDir = "${cliWorkspace}/src/github.com/jfrog/"
        jfrogCliDir = "${jfrogCliParentDir}jfrog-cli-go/"
        withEnv(["GO111MODULE=on","GOROOT=$goRoot","GOPATH=${cliWorkspace}","PATH+GOROOT=${goRoot}/bin", "JFROG_CLI_OFFER_CONFIG=false"]) {
            stage 'Go get'
            sh 'go version'
            sh "mkdir -p $jfrogCliParentDir"
            dir("$jfrogCliParentDir") {
                sh 'git clone https://github.com/jfrog/jfrog-cli-go.git'
                if (BRANCH?.trim()) {
                    dir('jfrog-cli-go') {
                        sh "git checkout $BRANCH"
                        dir('jfrog-cli/jfrog') {
                            sh 'go install'
                        }
                    }
                }
            }

            if ("$PUBLISH_NPM_PACKAGE".toBoolean()) {
                print "publishing npm package"
                publishNpmPackage()
            } else {
                // Publish to Bintray
                sh 'bin/jfrog --version > version'
                version = readFile('version').trim().split(" ")[2]
                print "publishing version: $version"
                for (int i = 0; i < architectures.size(); i++) {
                    def currentBuild = architectures[i]
                    stage "Build ${currentBuild.pkg}"
                    buildAndUpload(currentBuild.goos, currentBuild.goarch, currentBuild.pkg, currentBuild.fileExtention)
                }
            }
        }
    }
}

def uploadToBintray(pkg, fileName) {
    sh """#!/bin/bash
           bin/jfrog bt u $cliWorkspace/$fileName $subject/$repo/$pkg/$version /$version/$pkg/ --user=$USER_NAME --key=$KEY
        """
}

def buildAndUpload(goos, goarch, pkg, fileExtension) {
    def extension = fileExtension == null ? '' : fileExtension
    def fileName = "jfrog$fileExtension"
    dir("${jfrogCliDir}jfrog-cli/jfrog") {
        sh "env GOOS=$goos GOARCH=$goarch GO111MODULE=on go build"
        sh "mv $fileName $cliWorkspace"
    }

    uploadToBintray(pkg, fileName)
    sh "rm $fileName"
}

def publishNpmPackage() {
    dir ('${jfrogCliDir}npm/') {
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