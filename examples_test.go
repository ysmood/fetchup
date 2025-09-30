package fetchup_test

import (
	"fmt"
	"os"

	"github.com/ysmood/fetchup/pkg/golang_migrate"
	"github.com/ysmood/fetchup/pkg/golangci_lint"
)

func Example_install_golangci_lint() {
	os.Setenv("GOBIN", "bin")

	err := golangci_lint.Install()
	if err != nil {
		panic(err)
	}

	fmt.Print(shellExec("./bin/golangci-lint version --short"))

	// Output:
	// 2.5.0
}

func Example_install_golang_migrate() {
	os.Setenv("GOBIN", "bin")

	err := golang_migrate.Install()
	if err != nil {
		panic(err)
	}

	fmt.Print(shellExec("./bin/migrate -version"))

	// Output:
	// 4.19.0
}
