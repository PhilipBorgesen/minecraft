package profile_test

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/PhilipBorgesen/minecraft/profile"
)

// The following example shows how to retrieve and report all
// information about a Minecraft profile based on its username.
func Example() {

	// User to retrieve information about
	username := "nergalic"

	// Retrieve basic information
	p, err := profile.Load(username)
	if err != nil {

		// Handle error
		panic("failed to load: " + err.Error())
	}

	// Get case-corrected username and ID
	name := p.Name()
	id := p.ID()

	// Load previously used usernames
	hist, err := p.LoadNameHistory()
	if err != nil {

		// Handle error
		panic("failed to load name history: " + err.Error())
	}

	// Load cape, skin and model type
	props, err := p.LoadProperties()
	if err != nil {

		// Handle error
		panic("failed to load properties: " + err.Error())
	}

	// Get model type, skin and cape
	model := props.Model()
	skin, sOk := props.SkinURL()
	cape, cOk := props.CapeURL()

	// If profile has no custom skin
	if !sOk {

		skin = "<NONE>"
	}

	// If profile has no cape
	if !cOk {

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

// The following example shows how to retrieve a profile by ID
// and then save its skin to a .png file.
func ExampleProperties() {

	// Profile ID to retrieve skin for
	id := fetchProfileIdFromDatabase()

	// Load profile with skin information preloaded
	p, err := profile.LoadWithProperties(id)
	if err != nil {

		panic("failed to load profile: " + err.Error())
	}

	// We know properties already have been loaded, hence we
	// can use the Properties method instead of LoadProperties.
	rc, err := p.Properties().SkinReader()
	if err != nil {

		switch err.(type) {

		case profile.ErrNoSkin: // Profile had no skin.
			rc = readSteveDefault() // Fallback to default Steve skin.

		default: // Handle error
			panic("failed to load skin: " + err.Error())
		}
	}
	defer rc.Close()

	// Filename: <PROFILE_USERNAME>.png
	filename := p.Name() + ".png"

	bs, err := ioutil.ReadAll(rc)
	if err != nil || len(bs) == 0 {

		// Handle error
		panic("failed to load skin: " + err.Error())
	}

	if ioutil.WriteFile(filename, bs, 0666) != nil {

		// Handle error
		panic(fmt.Sprintf("failed to save skin to %s: %s", filename, err))
	}
}

func fetchProfileIdFromDatabase() string {

	// AxeLaw ID
	return "cabefc91b5df4c87886a6c604da2e46f"
}

func readSteveDefault() io.ReadCloser {

	panic("Test profile had no skin")
}
