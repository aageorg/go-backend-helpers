package utils_test

import (
	"fmt"
	"testing"

	"github.com/evgeniums/go-backend-helpers/pkg/test_utils"
)

func TestGlobalTestVars(t *testing.T) {

	fmt.Println(`Global testing flags can be set with ldflags when running tests like: go test -ldflags="-X 'github.com/evgeniums/go-backend-helpers/pkg/test_utils.Testing=true'"`)

	fmt.Printf("Current values:\n")
	fmt.Printf("test_utils.Testing=%v\n", test_utils.Testing)
	fmt.Printf("test_utils.ExternalConfigPath=%s\n", test_utils.ExternalConfigPath)
	fmt.Printf("test_utils.Phone=%v\n", test_utils.Phone)
}
