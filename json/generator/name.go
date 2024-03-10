package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"strings"
)

func NameTree() *Tree {
	return &Tree{
		Name: "Name",
		compare: func(r *Request) bool {
			last := r.LastName()
			return (strings.ToLower(last) == "name" || strings.HasSuffix(last, "Name")) &&
				(r.Schema.IsString() || r.Schema.IsAny())
		},
		resolve: func(r *Request) (interface{}, error) {
			index := gofakeit.Number(0, len(names)-1)
			return names[index], nil
		},
	}
}

var names = []string{
	"AuroraWaves",
	"BloomCrest",
	"CrystalVeil",
	"DewMeadow",
	"EchoForge",
	"FrostGuard",
	"HavenRoot",
	"IrisField",
	"JadeVista",
	"KaleidoSpace",
	"LunarFlare",
	"MysticPeak",
	"NebulaStream",
	"OrionTrail",
	"PulseNet",
	"QuartzBend",
	"RippleZone",
	"SolarBreeze",
	"TerraCove",
	"UltraQuest",
	"VortexEdge",
	"WillowSpark",
	"XenonPulse",
	"YieldPath",
	"ZenithLight",
	"AlphaSphere",
	"BetaBridge",
	"CirrusGate",
	"DeltaWave",
	"EclipseBound",
	"FlareCraft",
	"GroveNest",
	"HorizonKey",
	"InfinityLoop",
	"JoltForge",
	"KryptonGlow",
	"LumenHaven",
	"MirageStream",
	"NovaLink",
	"OasisDream",
	"PhantomRidge",
	"QuantumSea",
	"RiftValley",
	"SparkVenture",
	"TideHarbor",
	"UmbraPhase",
	"VertexField",
	"WhisperGlen",
	"ZephyrWing",
	"BlazeCraft",
	"CelestialPath",
	"DawnChaser",
	"EchoValley",
	"FrostForge",
	"GlimmerShore",
	"HorizonPeak",
	"InfinityBloom",
	"JadeVoyage",
	"KaleidoSky",
	"LunarHaven",
	"MysticGlade",
	"NebulaNest",
	"OrionQuest",
	"PrismPulse",
	"QuartzQuarry",
	"RavenRoost",
	"SolarFlare",
	"TideTreasure",
	"UmbraUnit",
	"VortexValley",
	"WhisperWoods",
	"XenonXylo",
	"YieldYarn",
	"ZephyrZone",
	"AetherArc",
	"BerylBay",
	"CrimsonCove",
	"DriftDream",
	"EclipseEdge",
	"FlareFountain",
	"GroveGuard",
	"HaloHarbor",
	"IrisIsle",
	"JasperJunction",
	"KarmaKey",
	"LumenLake",
	"MarbleMeadow",
	"NovaNiche",
	"OpalOasis",
	"PulsePoint",
	"QuiverQuill",
	"RiftRanger",
	"SparkSphere",
	"TerraTrove",
	"UtopiaUnfurl",
	"VividVale",
	"WillowWisp",
	"ZenithZing",
}
