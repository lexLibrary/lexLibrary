pipeline {
    agent any
    environment {
        GOPATH = '/go'
        REPO = '/go/src/github.com/lexLibrary/lexLibrary'
    }
    stages {
        stage('build') {
            agent {
                dockerfile { 
                    dir 'ci/build' 
                    args '-v $WORKSPACE:/go/src/github.com/lexLibrary/lexLibrary'
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
                            args '-v $WORKSPACE:/go/src/github.com/lexLibrary/lexLibrary'
                        }
                    }
                    steps {
                        sh '''
                            cd $REPO
                            go test  ./... -config ci/sqlite/config.yaml
                        '''
                    }
                }
            }
        }
    }
}

