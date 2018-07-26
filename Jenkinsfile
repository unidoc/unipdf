node {
    // Install the desired Go version
    def root = tool name: 'go 1.10.3', type: 'go'

    env.GOROOT="${root}"
    env.GOPATH="${WORKSPACE}/gopath"
    env.PATH="${root}/bin:${env.GOPATH}/bin:${env.PATH}"

    dir("${GOPATH}/src/github.com/unidoc/unidoc") {
        sh 'go version'

        stage('Checkout') {
            git url: 'https://github.com/unidoc/unidoc.git'
        }

        stage('Prepare') {
            // Get linter and other build tools.
            sh 'go get github.com/golang/lint/golint'
            sh 'go get github.com/tebeka/go2xunit'
            sh 'go get github.com/t-yuki/gocover-cobertura'

            // Get dependencies
            sh 'go get golang.org/x/image/tiff/lzw'
            sh 'go get github.com/boombuler/barcode'
        }

        stage('Linting') {
            // Go vet - List issues
            sh '(go vet ./... >govet.txt 2>&1) || true'

            // Go lint - List issues
            sh '(golint ./... >golint.txt 2>&1) || true'
        }


        stage('Testing') {
            // Go test - No tolerance.
            //sh 'go test -v ./... >gotest.txt 2>&1'
            sh '2>&1 go test -v ./... | tee gotest.txt'
        }

        stage('Test coverage') {
            sh 'go test -coverprofile=coverage.out ./...'
            sh 'gocover-cobertura < coverage.out > coverage.xml'
            step([$class: 'CoberturaPublisher', coberturaReportFile: 'coverage.xml'])
        }


        stage('Post') {
            // Assemble vet and lint info.
            warnings parserConfigurations: [
                [pattern: 'govet.txt', parserName: 'Go Vet'],
                [pattern: 'golint.txt', parserName: 'Go Lint']
            ]

            sh 'go2xunit -fail -input gotest.txt -output gotest.xml'
            junit "gotest.xml"
        }

    }
}