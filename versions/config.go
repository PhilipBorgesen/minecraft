package versions

import "github.com/PhilipBorgesen/minecraft/versions/internal"

// The endpoint to fetch version information from.
// For test purposes a cached response should be downloaded to testdata/cached/<SERVER PATH>
const versionsURL = internal.VersionsURL
