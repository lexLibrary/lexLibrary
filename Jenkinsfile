pipeline {
    agent any
    stages {
        stage('build') {
            agent {
                dockerfile { 
                    dir 'ci/build' 
                    args '-v $WORKSPACE:/go/src/github.com/lexLibrary/lexLibrary'
                }
            }
            environment {
                GOPATH = '/go'
                REPO = '/go/src/github.com/lexLibrary/lexLibrary'
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
        stage('lint and cover') {
            agent {
                dockerfile { 
                    dir 'ci/build' 
                    args '-v $WORKSPACE:/go/src/github.com/lexLibrary/lexLibrary'
                }
            }
            environment {
                GOPATH = '/go'
                REPO = '/go/src/github.com/lexLibrary/lexLibrary'
                HOME = '.'
            }
            steps {
                sh '''
                    cd $REPO
                    gometalinter ./... --vendor --deadline 1m --disable-all \
                        --enable=megacheck
                    go test -cover

                '''
            }
        }
        stage('test') {
            parallel {
                stage('sqlite') {
                    steps {
                        sh '''
                            cd ci
                            sh ./testDB.sh sqlite
                        '''
                    }
                }
                stage('postgres') {
                    steps {
                        sh '''
                            cd ci
                            sh ./testDB.sh postgres
                        '''
                    }
                }
                stage('mysql') {
                    steps {
                        sh '''
                            cd ci
                            sh ./testDB.sh mysql
                        '''
                    }
                }
                stage('cockroachdb') {
                    steps {
                        sh '''
                            cd ci
                            sh ./testDB.sh cockroachdb
                        '''
                    }
                }
                stage('tidb') {
                    steps {
                        sh '''
                            cd ci
                            sh ./testDB.sh tidb
                        '''
                    }
                }
            }
        }
    }
}

