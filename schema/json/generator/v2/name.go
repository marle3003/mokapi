package v2

import (
	"github.com/brianvoe/gofakeit/v6"
)

func newNameNode() *Node {
	return &Node{Name: "name", Fake: fakeName}
}

func fakeName(r *Request) (interface{}, error) {
	s := r.Schema
	var collection []string
	min := 0
	max := 12
	if s != nil && s.MinLength != nil {
		min = *s.MinLength
	}
	if s != nil && s.MaxLength != nil {
		max = *s.MaxLength
	}
	if min <= 3 && max >= 3 {
		collection = append(collection, names3...)
	}
	if min <= 4 && max >= 4 {
		collection = append(collection, names4...)
	}
	if min <= 5 && max >= 5 {
		collection = append(collection, names5...)
	}
	if min <= 6 && max >= 6 {
		collection = append(collection, names6...)
	}
	if max >= 12 {
		collection = append(collection, names...)
	}

	if len(collection) > 0 {
		index := gofakeit.Number(0, len(collection)-1)
		return collection[index], nil
	}
	return nil, NotSupported

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

var names3 = []string{"Zed", "Zen", "Lux", "Evo", "Vox", "Hex", "Arc", "Orb", "Neo", "Sol", "Ink", "Sky", "Kin", "Bio", "Eon", "Xis", "Ivy", "Jet"}
var names4 = []string{"Flux", "Fuse", "Halo", "Echo", "Nova", "Sync", "Aura", "Beam", "Axis", "Luna", "Apex", "Vibe", "Zeal"}
var names5 = []string{"Zephy", "Verve", "Nexus", "Zenix", "Focus", "Pulse", "Exalt", "Prism", "Vital", "Solis", "Evolve", "Quest", "Nova", "Zebra", "Unity", "Envis", "Axis", "Amity", "Lumin", "Swift"}
var names6 = []string{"Nebula", "Shadow", "Willow", "Arctic", "Comet", "Stellar", "Spirit", "Canyon", "Ember", "Horizon", "Jaguar", "Legend", "Meadow", "Phoenix", "Rocket", "Safari", "Silver", "Temple", "Utopia", "Velvet"}
