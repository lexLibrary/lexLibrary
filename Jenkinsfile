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
	    stage('megacheck') {
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
			     gometalinter ./... --vendor --concurrency 1 --deadline 30m --disable-all --enable=megacheck
			'''
		    }
	    }
	    stage('cover and race') {
                    steps {
			    sh '''
				cd ci
				sh ./testInDocker.sh static
			    '''
                    }
            }
        }
        stage('test databases') {
            parallel {
                stage('sqlite') {
                    steps {
                    sh '''
                        cd ci
                        sh ./testInDocker.sh sqlite
                    '''
                    }
                }
                stage('postgres') {
                    steps {
                    sh '''
                        cd ci
                        sh ./testInDocker.sh postgres
                    '''
                    }
                }
                stage('mysql') {
                    steps {
                    sh '''
                        cd ci
                        sh ./testInDocker.sh mysql
                    '''
                    }
                }
                stage('cockroachdb') {
                    steps {
                    sh '''
                        cd ci
                        sh ./testInDocker.sh cockroachdb
                    '''
                    }
                }
                stage('sqlserver') {
                    steps {
                    sh '''
                        cd ci
                        sh ./testInDocker.sh sqlserver
                    '''
                    }
                }
                stage('mariadb') {
                    steps {
                    sh '''
                        cd ci
                        sh ./testInDocker.sh mariadb
                    '''
                    }
                }
            }
        }
	stage('test browsers') {
            parallel {
                stage('firefox') {
                    steps {
                    sh '''
                        cd ci
                        sh ./testInDocker.sh firefox
                    '''
                    }
                }
		stage('chrome') {
                    steps {
                    sh '''
                        cd ci
                        sh ./testInDocker.sh chrome
                    '''
                    }
                }
	    }
	}
    }
}
