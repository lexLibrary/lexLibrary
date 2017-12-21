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
                    ./build.sh
                '''
            }
        }
        stage('static analysis') {
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
                    gometalinter ./... --vendor --deadline 5m --disable-all \
                        --enable=megacheck
                '''
                sh '''
                    cd $REPO
                    go test ./... -cover
                '''
                sh '''
                    cd $REPO
                    go test ./... -race
                '''
            }
        }
        stage('test') {
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
            stage('sqlserver') {
                steps {
                    sh '''
                        cd ci
                        sh ./testDB.sh sqlserver
                    '''
                }
            }
        }
    }
}

