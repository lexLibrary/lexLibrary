pipeline {
    agent any
    environment {
        GOPATH = '/go'
        REPO = '$GOPATH/src/github.com/lexLibrary/lexLibrary'
    }
    stages {
        stage('build') {
            agent {
                dockerfile { 
                    dir 'ci/build' 
                    args '-v $WORKSPACE:$REPO'
                }
            }
            environment {
                HOME = '.'
            }
            steps {
                sh '''
                    cd $REPO
                    go clean -i -a
                    go build -o lexLibrary
                '''
            }
        }
        stage('test') {
            parallel {
                stage('sqlite') {
                    agent {
                        dockerfile { 
                            dir 'ci/sqlite' 
                            args '-v $WORKSPACE:$REPO'
                        }
                    }
                    steps {
                        sh '''
                            cd $REPO
                            go test  ./...
                        '''
                    }
                }
            }
        }
    }
}

