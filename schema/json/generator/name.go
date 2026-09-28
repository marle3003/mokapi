package generator

import (
	"github.com/brianvoe/gofakeit/v7"
)

func newNameNode() *Node {
	return &Node{
		Name:       "name",
		Attributes: []string{"name"},
		Fake:       fakeName,
	}
}

func fakeName(r *Request) (any, error) {
	if v, ok := r.Context.Values["name"]; ok {
		return v, nil
	}

	s := r.Schema
	var collection []string
	minLength := 0
	maxLength := 12
	if s != nil && s.MinLength != nil {
		minLength = *s.MinLength
	}
	if s != nil && s.MaxLength != nil {
		maxLength = *s.MaxLength
	}
	if minLength <= 3 && maxLength >= 3 {
		collection = append(collection, names3...)
	}
	if minLength <= 4 && maxLength >= 4 {
		collection = append(collection, names4...)
	}
	if minLength <= 5 && maxLength >= 5 {
		collection = append(collection, names5...)
	}
	if minLength <= 6 && maxLength >= 6 {
		collection = append(collection, names6...)
	}
	if maxLength >= 12 {
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
