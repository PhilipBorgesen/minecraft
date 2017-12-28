package versions_test

import (
	"context"
	"log"

	"fmt"

	"github.com/PhilipBorgesen/minecraft/versions"
)

// The following example demonstrates how to retrieve an updated listing of
// Mincraft releases and print information for select versions.
func ExampleLoad() {
	ctx := context.TODO()

	vs, err := versions.Load(ctx)
	if err != nil {
		log.Fatal("Failed to fetch versions listing: " + err.Error())
	}

	fmt.Println("VERSION    TYPE     RELEASE DATE")
	fmt.Println("-------------------------------------------------")
	for _, s := range []string{"1.8.1", "b1.0", "rd-132211"} {
		v := vs.Versions[s]
		fmt.Printf("%-9s  %-7s  %s\n", v.ID, v.Type, v.Released.UTC())
	}
	// output:
	// VERSION    TYPE     RELEASE DATE
	// -------------------------------------------------
	// 1.8.1      release  2014-11-24 14:13:31 +0000 UTC
	// b1.0       beta     2010-12-19 22:00:00 +0000 UTC
	// rd-132211  alpha    2009-05-13 20:11:00 +0000 UTC
}
