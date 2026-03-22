// Package mockdata provides embedded mock JSON responses for all airline providers.
package mockdata

import _ "embed"

//go:embed garuda_indonesia_search_response.json
var Garuda []byte

//go:embed lion_air_search_response.json
var LionAir []byte

//go:embed batik_air_search_response.json
var BatikAir []byte

//go:embed airasia_search_response.json
var AirAsia []byte
