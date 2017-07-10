/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */
/*
 * The embedded character metrics specified in this file are distributed under the terms listed in
 * ./afms/MustRead.html.
 */

package fonts

import (
	"github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/model/textencoding"
)

// Font ZapfDingbats.  Implements Font interface.
// This is a built-in font and it is assumed that every reader has access to it.
type fontZapfDingbats struct {
	// By default encoder is not set, which means that we use the font's built in encoding.
	encoder textencoding.TextEncoder
}

func NewFontZapfDingbats() fontZapfDingbats {
	font := fontZapfDingbats{}
	return font
}

func (font fontZapfDingbats) SetEncoder(encoder textencoding.TextEncoder) {
	font.encoder = encoder
}

func (font fontZapfDingbats) GetGlyphCharMetrics(glyph string) (CharMetrics, bool) {
	metrics, has := zapfDingbatsCharMetrics[glyph]
	if !has {
		return metrics, false
	}

	return metrics, true
}

func (font fontZapfDingbats) ToPdfObject() core.PdfObject {
	obj := &core.PdfIndirectObject{}

	fontDict := core.MakeDict()
	fontDict.Set("Type", core.MakeName("Font"))
	fontDict.Set("Subtype", core.MakeName("Type1"))
	fontDict.Set("BaseFont", core.MakeName("ZapfDingbats"))
	if font.encoder != nil {
		fontDict.Set("Encoding", font.encoder.ToPdfObject())
	}

	obj.PdfObject = fontDict
	return obj
}

// ZapfDingbats font metics loaded from afms/ZapfDingbats.afm.  See afms/MustRead.html for license information.
var zapfDingbatsCharMetrics map[string]CharMetrics = map[string]CharMetrics{
	"a1":    {GlyphName: "a1", Wx: 974.000000, Wy: 0.000000},
	"a10":   {GlyphName: "a10", Wx: 692.000000, Wy: 0.000000},
	"a100":  {GlyphName: "a100", Wx: 668.000000, Wy: 0.000000},
	"a101":  {GlyphName: "a101", Wx: 732.000000, Wy: 0.000000},
	"a102":  {GlyphName: "a102", Wx: 544.000000, Wy: 0.000000},
	"a103":  {GlyphName: "a103", Wx: 544.000000, Wy: 0.000000},
	"a104":  {GlyphName: "a104", Wx: 910.000000, Wy: 0.000000},
	"a105":  {GlyphName: "a105", Wx: 911.000000, Wy: 0.000000},
	"a106":  {GlyphName: "a106", Wx: 667.000000, Wy: 0.000000},
	"a107":  {GlyphName: "a107", Wx: 760.000000, Wy: 0.000000},
	"a108":  {GlyphName: "a108", Wx: 760.000000, Wy: 0.000000},
	"a109":  {GlyphName: "a109", Wx: 626.000000, Wy: 0.000000},
	"a11":   {GlyphName: "a11", Wx: 960.000000, Wy: 0.000000},
	"a110":  {GlyphName: "a110", Wx: 694.000000, Wy: 0.000000},
	"a111":  {GlyphName: "a111", Wx: 595.000000, Wy: 0.000000},
	"a112":  {GlyphName: "a112", Wx: 776.000000, Wy: 0.000000},
	"a117":  {GlyphName: "a117", Wx: 690.000000, Wy: 0.000000},
	"a118":  {GlyphName: "a118", Wx: 791.000000, Wy: 0.000000},
	"a119":  {GlyphName: "a119", Wx: 790.000000, Wy: 0.000000},
	"a12":   {GlyphName: "a12", Wx: 939.000000, Wy: 0.000000},
	"a120":  {GlyphName: "a120", Wx: 788.000000, Wy: 0.000000},
	"a121":  {GlyphName: "a121", Wx: 788.000000, Wy: 0.000000},
	"a122":  {GlyphName: "a122", Wx: 788.000000, Wy: 0.000000},
	"a123":  {GlyphName: "a123", Wx: 788.000000, Wy: 0.000000},
	"a124":  {GlyphName: "a124", Wx: 788.000000, Wy: 0.000000},
	"a125":  {GlyphName: "a125", Wx: 788.000000, Wy: 0.000000},
	"a126":  {GlyphName: "a126", Wx: 788.000000, Wy: 0.000000},
	"a127":  {GlyphName: "a127", Wx: 788.000000, Wy: 0.000000},
	"a128":  {GlyphName: "a128", Wx: 788.000000, Wy: 0.000000},
	"a129":  {GlyphName: "a129", Wx: 788.000000, Wy: 0.000000},
	"a13":   {GlyphName: "a13", Wx: 549.000000, Wy: 0.000000},
	"a130":  {GlyphName: "a130", Wx: 788.000000, Wy: 0.000000},
	"a131":  {GlyphName: "a131", Wx: 788.000000, Wy: 0.000000},
	"a132":  {GlyphName: "a132", Wx: 788.000000, Wy: 0.000000},
	"a133":  {GlyphName: "a133", Wx: 788.000000, Wy: 0.000000},
	"a134":  {GlyphName: "a134", Wx: 788.000000, Wy: 0.000000},
	"a135":  {GlyphName: "a135", Wx: 788.000000, Wy: 0.000000},
	"a136":  {GlyphName: "a136", Wx: 788.000000, Wy: 0.000000},
	"a137":  {GlyphName: "a137", Wx: 788.000000, Wy: 0.000000},
	"a138":  {GlyphName: "a138", Wx: 788.000000, Wy: 0.000000},
	"a139":  {GlyphName: "a139", Wx: 788.000000, Wy: 0.000000},
	"a14":   {GlyphName: "a14", Wx: 855.000000, Wy: 0.000000},
	"a140":  {GlyphName: "a140", Wx: 788.000000, Wy: 0.000000},
	"a141":  {GlyphName: "a141", Wx: 788.000000, Wy: 0.000000},
	"a142":  {GlyphName: "a142", Wx: 788.000000, Wy: 0.000000},
	"a143":  {GlyphName: "a143", Wx: 788.000000, Wy: 0.000000},
	"a144":  {GlyphName: "a144", Wx: 788.000000, Wy: 0.000000},
	"a145":  {GlyphName: "a145", Wx: 788.000000, Wy: 0.000000},
	"a146":  {GlyphName: "a146", Wx: 788.000000, Wy: 0.000000},
	"a147":  {GlyphName: "a147", Wx: 788.000000, Wy: 0.000000},
	"a148":  {GlyphName: "a148", Wx: 788.000000, Wy: 0.000000},
	"a149":  {GlyphName: "a149", Wx: 788.000000, Wy: 0.000000},
	"a15":   {GlyphName: "a15", Wx: 911.000000, Wy: 0.000000},
	"a150":  {GlyphName: "a150", Wx: 788.000000, Wy: 0.000000},
	"a151":  {GlyphName: "a151", Wx: 788.000000, Wy: 0.000000},
	"a152":  {GlyphName: "a152", Wx: 788.000000, Wy: 0.000000},
	"a153":  {GlyphName: "a153", Wx: 788.000000, Wy: 0.000000},
	"a154":  {GlyphName: "a154", Wx: 788.000000, Wy: 0.000000},
	"a155":  {GlyphName: "a155", Wx: 788.000000, Wy: 0.000000},
	"a156":  {GlyphName: "a156", Wx: 788.000000, Wy: 0.000000},
	"a157":  {GlyphName: "a157", Wx: 788.000000, Wy: 0.000000},
	"a158":  {GlyphName: "a158", Wx: 788.000000, Wy: 0.000000},
	"a159":  {GlyphName: "a159", Wx: 788.000000, Wy: 0.000000},
	"a16":   {GlyphName: "a16", Wx: 933.000000, Wy: 0.000000},
	"a160":  {GlyphName: "a160", Wx: 894.000000, Wy: 0.000000},
	"a161":  {GlyphName: "a161", Wx: 838.000000, Wy: 0.000000},
	"a162":  {GlyphName: "a162", Wx: 924.000000, Wy: 0.000000},
	"a163":  {GlyphName: "a163", Wx: 1016.000000, Wy: 0.000000},
	"a164":  {GlyphName: "a164", Wx: 458.000000, Wy: 0.000000},
	"a165":  {GlyphName: "a165", Wx: 924.000000, Wy: 0.000000},
	"a166":  {GlyphName: "a166", Wx: 918.000000, Wy: 0.000000},
	"a167":  {GlyphName: "a167", Wx: 927.000000, Wy: 0.000000},
	"a168":  {GlyphName: "a168", Wx: 928.000000, Wy: 0.000000},
	"a169":  {GlyphName: "a169", Wx: 928.000000, Wy: 0.000000},
	"a17":   {GlyphName: "a17", Wx: 945.000000, Wy: 0.000000},
	"a170":  {GlyphName: "a170", Wx: 834.000000, Wy: 0.000000},
	"a171":  {GlyphName: "a171", Wx: 873.000000, Wy: 0.000000},
	"a172":  {GlyphName: "a172", Wx: 828.000000, Wy: 0.000000},
	"a173":  {GlyphName: "a173", Wx: 924.000000, Wy: 0.000000},
	"a174":  {GlyphName: "a174", Wx: 917.000000, Wy: 0.000000},
	"a175":  {GlyphName: "a175", Wx: 930.000000, Wy: 0.000000},
	"a176":  {GlyphName: "a176", Wx: 931.000000, Wy: 0.000000},
	"a177":  {GlyphName: "a177", Wx: 463.000000, Wy: 0.000000},
	"a178":  {GlyphName: "a178", Wx: 883.000000, Wy: 0.000000},
	"a179":  {GlyphName: "a179", Wx: 836.000000, Wy: 0.000000},
	"a18":   {GlyphName: "a18", Wx: 974.000000, Wy: 0.000000},
	"a180":  {GlyphName: "a180", Wx: 867.000000, Wy: 0.000000},
	"a181":  {GlyphName: "a181", Wx: 696.000000, Wy: 0.000000},
	"a182":  {GlyphName: "a182", Wx: 874.000000, Wy: 0.000000},
	"a183":  {GlyphName: "a183", Wx: 760.000000, Wy: 0.000000},
	"a184":  {GlyphName: "a184", Wx: 946.000000, Wy: 0.000000},
	"a185":  {GlyphName: "a185", Wx: 865.000000, Wy: 0.000000},
	"a186":  {GlyphName: "a186", Wx: 967.000000, Wy: 0.000000},
	"a187":  {GlyphName: "a187", Wx: 831.000000, Wy: 0.000000},
	"a188":  {GlyphName: "a188", Wx: 873.000000, Wy: 0.000000},
	"a189":  {GlyphName: "a189", Wx: 927.000000, Wy: 0.000000},
	"a19":   {GlyphName: "a19", Wx: 755.000000, Wy: 0.000000},
	"a190":  {GlyphName: "a190", Wx: 970.000000, Wy: 0.000000},
	"a191":  {GlyphName: "a191", Wx: 918.000000, Wy: 0.000000},
	"a192":  {GlyphName: "a192", Wx: 748.000000, Wy: 0.000000},
	"a193":  {GlyphName: "a193", Wx: 836.000000, Wy: 0.000000},
	"a194":  {GlyphName: "a194", Wx: 771.000000, Wy: 0.000000},
	"a195":  {GlyphName: "a195", Wx: 888.000000, Wy: 0.000000},
	"a196":  {GlyphName: "a196", Wx: 748.000000, Wy: 0.000000},
	"a197":  {GlyphName: "a197", Wx: 771.000000, Wy: 0.000000},
	"a198":  {GlyphName: "a198", Wx: 888.000000, Wy: 0.000000},
	"a199":  {GlyphName: "a199", Wx: 867.000000, Wy: 0.000000},
	"a2":    {GlyphName: "a2", Wx: 961.000000, Wy: 0.000000},
	"a20":   {GlyphName: "a20", Wx: 846.000000, Wy: 0.000000},
	"a200":  {GlyphName: "a200", Wx: 696.000000, Wy: 0.000000},
	"a201":  {GlyphName: "a201", Wx: 874.000000, Wy: 0.000000},
	"a202":  {GlyphName: "a202", Wx: 974.000000, Wy: 0.000000},
	"a203":  {GlyphName: "a203", Wx: 762.000000, Wy: 0.000000},
	"a204":  {GlyphName: "a204", Wx: 759.000000, Wy: 0.000000},
	"a205":  {GlyphName: "a205", Wx: 509.000000, Wy: 0.000000},
	"a206":  {GlyphName: "a206", Wx: 410.000000, Wy: 0.000000},
	"a21":   {GlyphName: "a21", Wx: 762.000000, Wy: 0.000000},
	"a22":   {GlyphName: "a22", Wx: 761.000000, Wy: 0.000000},
	"a23":   {GlyphName: "a23", Wx: 571.000000, Wy: 0.000000},
	"a24":   {GlyphName: "a24", Wx: 677.000000, Wy: 0.000000},
	"a25":   {GlyphName: "a25", Wx: 763.000000, Wy: 0.000000},
	"a26":   {GlyphName: "a26", Wx: 760.000000, Wy: 0.000000},
	"a27":   {GlyphName: "a27", Wx: 759.000000, Wy: 0.000000},
	"a28":   {GlyphName: "a28", Wx: 754.000000, Wy: 0.000000},
	"a29":   {GlyphName: "a29", Wx: 786.000000, Wy: 0.000000},
	"a3":    {GlyphName: "a3", Wx: 980.000000, Wy: 0.000000},
	"a30":   {GlyphName: "a30", Wx: 788.000000, Wy: 0.000000},
	"a31":   {GlyphName: "a31", Wx: 788.000000, Wy: 0.000000},
	"a32":   {GlyphName: "a32", Wx: 790.000000, Wy: 0.000000},
	"a33":   {GlyphName: "a33", Wx: 793.000000, Wy: 0.000000},
	"a34":   {GlyphName: "a34", Wx: 794.000000, Wy: 0.000000},
	"a35":   {GlyphName: "a35", Wx: 816.000000, Wy: 0.000000},
	"a36":   {GlyphName: "a36", Wx: 823.000000, Wy: 0.000000},
	"a37":   {GlyphName: "a37", Wx: 789.000000, Wy: 0.000000},
	"a38":   {GlyphName: "a38", Wx: 841.000000, Wy: 0.000000},
	"a39":   {GlyphName: "a39", Wx: 823.000000, Wy: 0.000000},
	"a4":    {GlyphName: "a4", Wx: 719.000000, Wy: 0.000000},
	"a40":   {GlyphName: "a40", Wx: 833.000000, Wy: 0.000000},
	"a41":   {GlyphName: "a41", Wx: 816.000000, Wy: 0.000000},
	"a42":   {GlyphName: "a42", Wx: 831.000000, Wy: 0.000000},
	"a43":   {GlyphName: "a43", Wx: 923.000000, Wy: 0.000000},
	"a44":   {GlyphName: "a44", Wx: 744.000000, Wy: 0.000000},
	"a45":   {GlyphName: "a45", Wx: 723.000000, Wy: 0.000000},
	"a46":   {GlyphName: "a46", Wx: 749.000000, Wy: 0.000000},
	"a47":   {GlyphName: "a47", Wx: 790.000000, Wy: 0.000000},
	"a48":   {GlyphName: "a48", Wx: 792.000000, Wy: 0.000000},
	"a49":   {GlyphName: "a49", Wx: 695.000000, Wy: 0.000000},
	"a5":    {GlyphName: "a5", Wx: 789.000000, Wy: 0.000000},
	"a50":   {GlyphName: "a50", Wx: 776.000000, Wy: 0.000000},
	"a51":   {GlyphName: "a51", Wx: 768.000000, Wy: 0.000000},
	"a52":   {GlyphName: "a52", Wx: 792.000000, Wy: 0.000000},
	"a53":   {GlyphName: "a53", Wx: 759.000000, Wy: 0.000000},
	"a54":   {GlyphName: "a54", Wx: 707.000000, Wy: 0.000000},
	"a55":   {GlyphName: "a55", Wx: 708.000000, Wy: 0.000000},
	"a56":   {GlyphName: "a56", Wx: 682.000000, Wy: 0.000000},
	"a57":   {GlyphName: "a57", Wx: 701.000000, Wy: 0.000000},
	"a58":   {GlyphName: "a58", Wx: 826.000000, Wy: 0.000000},
	"a59":   {GlyphName: "a59", Wx: 815.000000, Wy: 0.000000},
	"a6":    {GlyphName: "a6", Wx: 494.000000, Wy: 0.000000},
	"a60":   {GlyphName: "a60", Wx: 789.000000, Wy: 0.000000},
	"a61":   {GlyphName: "a61", Wx: 789.000000, Wy: 0.000000},
	"a62":   {GlyphName: "a62", Wx: 707.000000, Wy: 0.000000},
	"a63":   {GlyphName: "a63", Wx: 687.000000, Wy: 0.000000},
	"a64":   {GlyphName: "a64", Wx: 696.000000, Wy: 0.000000},
	"a65":   {GlyphName: "a65", Wx: 689.000000, Wy: 0.000000},
	"a66":   {GlyphName: "a66", Wx: 786.000000, Wy: 0.000000},
	"a67":   {GlyphName: "a67", Wx: 787.000000, Wy: 0.000000},
	"a68":   {GlyphName: "a68", Wx: 713.000000, Wy: 0.000000},
	"a69":   {GlyphName: "a69", Wx: 791.000000, Wy: 0.000000},
	"a7":    {GlyphName: "a7", Wx: 552.000000, Wy: 0.000000},
	"a70":   {GlyphName: "a70", Wx: 785.000000, Wy: 0.000000},
	"a71":   {GlyphName: "a71", Wx: 791.000000, Wy: 0.000000},
	"a72":   {GlyphName: "a72", Wx: 873.000000, Wy: 0.000000},
	"a73":   {GlyphName: "a73", Wx: 761.000000, Wy: 0.000000},
	"a74":   {GlyphName: "a74", Wx: 762.000000, Wy: 0.000000},
	"a75":   {GlyphName: "a75", Wx: 759.000000, Wy: 0.000000},
	"a76":   {GlyphName: "a76", Wx: 892.000000, Wy: 0.000000},
	"a77":   {GlyphName: "a77", Wx: 892.000000, Wy: 0.000000},
	"a78":   {GlyphName: "a78", Wx: 788.000000, Wy: 0.000000},
	"a79":   {GlyphName: "a79", Wx: 784.000000, Wy: 0.000000},
	"a8":    {GlyphName: "a8", Wx: 537.000000, Wy: 0.000000},
	"a81":   {GlyphName: "a81", Wx: 438.000000, Wy: 0.000000},
	"a82":   {GlyphName: "a82", Wx: 138.000000, Wy: 0.000000},
	"a83":   {GlyphName: "a83", Wx: 277.000000, Wy: 0.000000},
	"a84":   {GlyphName: "a84", Wx: 415.000000, Wy: 0.000000},
	"a85":   {GlyphName: "a85", Wx: 509.000000, Wy: 0.000000},
	"a86":   {GlyphName: "a86", Wx: 410.000000, Wy: 0.000000},
	"a87":   {GlyphName: "a87", Wx: 234.000000, Wy: 0.000000},
	"a88":   {GlyphName: "a88", Wx: 234.000000, Wy: 0.000000},
	"a89":   {GlyphName: "a89", Wx: 390.000000, Wy: 0.000000},
	"a9":    {GlyphName: "a9", Wx: 577.000000, Wy: 0.000000},
	"a90":   {GlyphName: "a90", Wx: 390.000000, Wy: 0.000000},
	"a91":   {GlyphName: "a91", Wx: 276.000000, Wy: 0.000000},
	"a92":   {GlyphName: "a92", Wx: 276.000000, Wy: 0.000000},
	"a93":   {GlyphName: "a93", Wx: 317.000000, Wy: 0.000000},
	"a94":   {GlyphName: "a94", Wx: 317.000000, Wy: 0.000000},
	"a95":   {GlyphName: "a95", Wx: 334.000000, Wy: 0.000000},
	"a96":   {GlyphName: "a96", Wx: 334.000000, Wy: 0.000000},
	"a97":   {GlyphName: "a97", Wx: 392.000000, Wy: 0.000000},
	"a98":   {GlyphName: "a98", Wx: 392.000000, Wy: 0.000000},
	"a99":   {GlyphName: "a99", Wx: 668.000000, Wy: 0.000000},
	"space": {GlyphName: "space", Wx: 278.000000, Wy: 0.000000},
}
