package profile_test

import (
	"fmt"
	"log"

	"context"
	"github.com/PhilipBorgesen/minecraft/profile"
)

// The following example shows how to retrieve and report all
// information about a Minecraft profile based on its username.
func Example() {
	ctx := context.TODO()

	// User to retrieve information about
	const username = "nergalic"

	// Retrieve basic information
	p, err := profile.Load(ctx, username)
	if err != nil {
		log.Fatalf("failed to load: %s", err)
	}

	// Get case-corrected username and ID
	name, id := p.Name, p.ID

	// Load previously associated usernames
	hist, err := p.LoadNameHistory(ctx, false)
	if err != nil {
		log.Fatalf("failed to load name history: %s", err)
	}

	// Load cape, skin and model type
	props, err := p.LoadProperties(ctx, false)
	if err != nil {
		log.Fatalf("failed to load properties: %s", err)
	}

	// Get model type, skin and cape
	model, skin, cape := props.Model, props.SkinURL, props.CapeURL

	// If profile has no custom skin
	if skin == "" {
		skin = "<NONE>"
	}

	// If profile has no cape
	if cape == "" {
		cape = "<NONE>"
	}

	// Report information
	fmt.Printf("INFORMATION FOR:         %32s\n", username)
	fmt.Println("---------------------------------------------------------")
	fmt.Printf("CASE-CORRECTED USERNAME: %32s\n", name)
	fmt.Printf("ID:                      %32s\n", id)
	fmt.Printf("PRIOR NAMES:             %32s\n", fmt.Sprint(hist))
	fmt.Println()
	fmt.Printf("SKIN MODEL:              %32s\n", model)
	fmt.Printf("SKIN URL:                %32s\n", skin)
	fmt.Printf("CAPE URL:                %32s\n", cape)

	// Example output:
	//
	// INFORMATION FOR:                                 nergalic
	// ---------------------------------------------------------
	// CASE-CORRECTED USERNAME:                         Nergalic
	// ID:                      087cc153c3434ff7ac497de1569affa1
	// PRIOR NAMES:                              [GeneralSezuan]
	//
	// SKIN MODEL:                                         Steve
	// SKIN URL:                http://textures.minecraft.net/texture/5b40f251f7c8db60943495db6bf54353102d6cad20d2299d5f973f36b4f3677e
	// CAPE URL:                                          <NONE>
}
