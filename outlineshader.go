//go:build ignore

//kage:unit pixels

// Package zen is the root for all ebiten-zen files
package zen

var OutlineThickness float
var OutlineColor vec4

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	var ret vec4

	// this could be used as a shadow
	// ret.r = OutlineColor.r * imageSrc0At(srcPos+vec2(2, 2)).a
	// ret.g = OutlineColor.g * imageSrc0At(srcPos+vec2(2, 2)).a
	// ret.b = OutlineColor.b * imageSrc0At(srcPos+vec2(2, 2)).a
	// ret.a = OutlineColor.a * imageSrc0At(srcPos+vec2(2, 2)).a

	o := OutlineThickness
	ret = max(max(max(OutlineColor*imageSrc0At(srcPos+vec2(o, o)).a,
		OutlineColor*imageSrc0At(srcPos+vec2(o, -o)).a),
		max(OutlineColor*imageSrc0At(srcPos+vec2(-o, o)).a,
			OutlineColor*imageSrc0At(srcPos+vec2(-o, -o)).a)),
		max(max(OutlineColor*imageSrc0At(srcPos+vec2(o, 0)).a,
			OutlineColor*imageSrc0At(srcPos+vec2(-o, 0)).a),
			max(OutlineColor*imageSrc0At(srcPos+vec2(0, o)).a,
				OutlineColor*imageSrc0At(srcPos+vec2(0, -o)).a)))

	// ret.r = max(max(OutlineColor.r*imageSrc0At(srcPos+vec2(2, -2)).a, OutlineColor.r*imageSrc0At(srcPos+vec2(2, 2)).a),
	// 	max(OutlineColor.r*imageSrc0At(srcPos+vec2(-2, -2)).a, OutlineColor.r*imageSrc0At(srcPos+vec2(-2, 2)).a))
	// ret.g = max(max(OutlineColor.g*imageSrc0At(srcPos+vec2(2, -2)).a, OutlineColor.g*imageSrc0At(srcPos+vec2(2, 2)).a),
	// 	max(OutlineColor.g*imageSrc0At(srcPos+vec2(-2, -2)).a, OutlineColor.g*imageSrc0At(srcPos+vec2(-2, 2)).a))
	// ret.b = max(max(OutlineColor.b*imageSrc0At(srcPos+vec2(2, -2)).a, OutlineColor.b*imageSrc0At(srcPos+vec2(2, 2)).a),
	// 	max(OutlineColor.b*imageSrc0At(srcPos+vec2(-2, -2)).a, OutlineColor.b*imageSrc0At(srcPos+vec2(-2, 2)).a))
	// ret.a = max(max(OutlineColor.a*imageSrc0At(srcPos+vec2(2, -2)).a, OutlineColor.a*imageSrc0At(srcPos+vec2(2, 2)).a),
	// 	max(OutlineColor.a*imageSrc0At(srcPos+vec2(-2, -2)).a, OutlineColor.a*imageSrc0At(srcPos+vec2(-2, 2)).a))

	if imageSrc0At(srcPos).a > 0 {
		ret.r = imageSrc0At(srcPos).r
		ret.g = imageSrc0At(srcPos).g
		ret.b = imageSrc0At(srcPos).b
		ret.a = imageSrc0At(srcPos).a
	}
	return ret
}
