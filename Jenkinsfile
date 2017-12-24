pipeline {
    agent any
    options { 
        disableConcurrentBuilds() 
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
        stage('test sqlite') {
            steps {
                sh '''
                    cd ci
                    sh ./testDB.sh sqlite
                '''
            }
        }
        stage('test postgres') {
            steps {
                sh '''
                    cd ci
                    sh ./testDB.sh postgres
                '''
            }
        }
        stage('test mysql') {
            steps {
                sh '''
                    cd ci
                    sh ./testDB.sh mysql
                '''
            }
        }
        stage('test cockroachdb') {
            steps {
                sh '''
                    cd ci
                    sh ./testDB.sh cockroachdb
                '''
            }
        }
        stage('test tidb') {
            steps {
                sh '''
                    cd ci
                    sh ./testDB.sh tidb
                '''
            }
        }
        stage('test sqlserver') {
            steps {
                sh '''
                    cd ci
                    sh ./testDB.sh sqlserver
                '''
            }
        }
    }
}