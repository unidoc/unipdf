node {
    // Install the desired Go version
    def root = tool name: 'go 1.10.3', type: 'go'

    env.GOROOT="${root}"
    env.GOPATH="${WORKSPACE}/gopath"
    env.PATH="${root}/bin:${env.GOPATH}/bin:${env.PATH}"

    dir("${GOPATH}/src/github.com/unidoc/unidoc") {
        sh 'go version'

        stage('Checkout') {
            echo "Pulling unidoc on branch ${env.BRANCH_NAME}"
            checkout scm
        }

        stage('Prepare') {
            // Get linter and other build tools.
            sh 'go get -u golang.org/x/lint/golint'
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

    dir("${GOPATH}/src/github.com/unidoc/unidoc-examples") {
        stage('Build examples') {
            // Pull unidoc-examples from connected branch, or master otherwise.
            def examplesBranch = "master"
            switch("${env.BRANCH_NAME}") {
                case "v3":
                    examplesBranch = "v3"
                    break
            }
            echo "Pulling unidoc-examples on branch ${examplesBranch}"
            git url: 'https://github.com/unidoc/unidoc-examples.git', branch: examplesBranch
            
            // Dependencies for examples.
            sh 'go get github.com/wcharczuk/go-chart'

            // Build all examples.
            sh 'find . -name "*.go" -print0 | xargs -0 -n1 go build'
        }

        stage('Passthrough benchmark pdfdb_small') {
            sh './pdf_passthrough_bench /home/jenkins/corpus/pdfdb_small/* | grep -v "Testing " | grep -v "copy of" | grep -v "To get " | grep -v " - pass"'
        }
    }
}