package profile_test

import (
	"fmt"
	"log"

	"context"

	"time"

	"github.com/PhilipBorgesen/minecraft/profile"
)

// The following example shows how to retrieve and report all information about
// a Minecraft profile based on its username.
func Example() {
	ctx := context.TODO()

	// User to retrieve information about
	const username = "nergalic"

	// Retrieve basic information
	p, err := profile.Load(ctx, username)
	if err != nil {
		log.Fatalf("Failed to load profile: %s", err)
	}

	// Get case-corrected username and ID
	name, id := p.Name, p.ID

	// Load previously associated usernames
	hist, err := p.LoadNameHistory(ctx, false)
	if err != nil {
		log.Fatalf("Failed to load profile name history: %s", err)
	}

	// Load cape, skin and model type
	props, err := p.LoadProperties(ctx, false)
	if err != nil {
		log.Fatalf("Failed to load profile properties: %s", err)
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
	// output:
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

// The following example shows how to retrieve a Minecraft profile based on the
// username it was registered under.
func ExampleLoadAtTime() {
	ctx := context.TODO()

	// Username the profile originally used
	const origUser = "GeneralSezuan"

	// Retrieve basic information of profile
	p, err := profile.LoadAtTime(ctx, origUser, time.Unix(0, 0))
	if err != nil {
		log.Fatalf("Failed to load profile: %s", err)
	}

	fmt.Println("Current username: " + p.Name)
	fmt.Println("      Profile ID: " + p.ID)
	// output:
	// Current username: Nergalic
	//       Profile ID: 087cc153c3434ff7ac497de1569affa1
}

// The following example demonstrates how to load basic information of multiple
// Minecraft profiles based on their currently associated usernames.
func ExampleLoadMany() {
	ctx := context.TODO()

	// Retrieve basic information of profiles that exist
	ps, err := profile.LoadMany(ctx, "nergalic", "breesakana", "doesNotExist")
	if err != nil {
		log.Fatalf("Failed to load profiles: %s", err)
	}

	for _, p := range ps {
		fmt.Printf("%-10s %s\n", p.Name, p.ID)
	}
	// Unordered output:
	// BreeSakana d9a5b542ce88442aaab38ec13e6c7773
	// Nergalic   087cc153c3434ff7ac497de1569affa1
}
