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

// Font Symbol.  Implements Font interface.
// This is a built-in font and it is assumed that every reader has access to it.
type fontSymbol struct {
	// By default encoder is not set, which means that we use the font's built in encoding.
	encoder textencoding.TextEncoder
}

func NewFontSymbol() fontSymbol {
	font := fontSymbol{}
	return font
}

func (font fontSymbol) SetEncoder(encoder textencoding.TextEncoder) {
	font.encoder = encoder
}

func (font fontSymbol) GetGlyphCharMetrics(glyph string) (CharMetrics, bool) {
	metrics, has := symbolCharMetrics[glyph]
	if !has {
		return metrics, false
	}

	return metrics, true
}

func (font fontSymbol) ToPdfObject() core.PdfObject {
	obj := &core.PdfIndirectObject{}

	fontDict := core.MakeDict()
	fontDict.Set("Type", core.MakeName("Font"))
	fontDict.Set("Subtype", core.MakeName("Type1"))
	fontDict.Set("BaseFont", core.MakeName("Symbol"))
	if font.encoder != nil {
		fontDict.Set("Encoding", font.encoder.ToPdfObject())
	}

	obj.PdfObject = fontDict
	return obj
}

// Symbol font metics loaded from afms/Symbol.afm.  See afms/MustRead.html for license information.
var symbolCharMetrics map[string]CharMetrics = map[string]CharMetrics{
	"Alpha":          {GlyphName: "Alpha", Wx: 722.000000, Wy: 0.000000},
	"Beta":           {GlyphName: "Beta", Wx: 667.000000, Wy: 0.000000},
	"Chi":            {GlyphName: "Chi", Wx: 722.000000, Wy: 0.000000},
	"Delta":          {GlyphName: "Delta", Wx: 612.000000, Wy: 0.000000},
	"Epsilon":        {GlyphName: "Epsilon", Wx: 611.000000, Wy: 0.000000},
	"Eta":            {GlyphName: "Eta", Wx: 722.000000, Wy: 0.000000},
	"Euro":           {GlyphName: "Euro", Wx: 750.000000, Wy: 0.000000},
	"Gamma":          {GlyphName: "Gamma", Wx: 603.000000, Wy: 0.000000},
	"Ifraktur":       {GlyphName: "Ifraktur", Wx: 686.000000, Wy: 0.000000},
	"Iota":           {GlyphName: "Iota", Wx: 333.000000, Wy: 0.000000},
	"Kappa":          {GlyphName: "Kappa", Wx: 722.000000, Wy: 0.000000},
	"Lambda":         {GlyphName: "Lambda", Wx: 686.000000, Wy: 0.000000},
	"Mu":             {GlyphName: "Mu", Wx: 889.000000, Wy: 0.000000},
	"Nu":             {GlyphName: "Nu", Wx: 722.000000, Wy: 0.000000},
	"Omega":          {GlyphName: "Omega", Wx: 768.000000, Wy: 0.000000},
	"Omicron":        {GlyphName: "Omicron", Wx: 722.000000, Wy: 0.000000},
	"Phi":            {GlyphName: "Phi", Wx: 763.000000, Wy: 0.000000},
	"Pi":             {GlyphName: "Pi", Wx: 768.000000, Wy: 0.000000},
	"Psi":            {GlyphName: "Psi", Wx: 795.000000, Wy: 0.000000},
	"Rfraktur":       {GlyphName: "Rfraktur", Wx: 795.000000, Wy: 0.000000},
	"Rho":            {GlyphName: "Rho", Wx: 556.000000, Wy: 0.000000},
	"Sigma":          {GlyphName: "Sigma", Wx: 592.000000, Wy: 0.000000},
	"Tau":            {GlyphName: "Tau", Wx: 611.000000, Wy: 0.000000},
	"Theta":          {GlyphName: "Theta", Wx: 741.000000, Wy: 0.000000},
	"Upsilon":        {GlyphName: "Upsilon", Wx: 690.000000, Wy: 0.000000},
	"Upsilon1":       {GlyphName: "Upsilon1", Wx: 620.000000, Wy: 0.000000},
	"Xi":             {GlyphName: "Xi", Wx: 645.000000, Wy: 0.000000},
	"Zeta":           {GlyphName: "Zeta", Wx: 611.000000, Wy: 0.000000},
	"aleph":          {GlyphName: "aleph", Wx: 823.000000, Wy: 0.000000},
	"alpha":          {GlyphName: "alpha", Wx: 631.000000, Wy: 0.000000},
	"ampersand":      {GlyphName: "ampersand", Wx: 778.000000, Wy: 0.000000},
	"angle":          {GlyphName: "angle", Wx: 768.000000, Wy: 0.000000},
	"angleleft":      {GlyphName: "angleleft", Wx: 329.000000, Wy: 0.000000},
	"angleright":     {GlyphName: "angleright", Wx: 329.000000, Wy: 0.000000},
	"apple":          {GlyphName: "apple", Wx: 790.000000, Wy: 0.000000},
	"approxequal":    {GlyphName: "approxequal", Wx: 549.000000, Wy: 0.000000},
	"arrowboth":      {GlyphName: "arrowboth", Wx: 1042.000000, Wy: 0.000000},
	"arrowdblboth":   {GlyphName: "arrowdblboth", Wx: 1042.000000, Wy: 0.000000},
	"arrowdbldown":   {GlyphName: "arrowdbldown", Wx: 603.000000, Wy: 0.000000},
	"arrowdblleft":   {GlyphName: "arrowdblleft", Wx: 987.000000, Wy: 0.000000},
	"arrowdblright":  {GlyphName: "arrowdblright", Wx: 987.000000, Wy: 0.000000},
	"arrowdblup":     {GlyphName: "arrowdblup", Wx: 603.000000, Wy: 0.000000},
	"arrowdown":      {GlyphName: "arrowdown", Wx: 603.000000, Wy: 0.000000},
	"arrowhorizex":   {GlyphName: "arrowhorizex", Wx: 1000.000000, Wy: 0.000000},
	"arrowleft":      {GlyphName: "arrowleft", Wx: 987.000000, Wy: 0.000000},
	"arrowright":     {GlyphName: "arrowright", Wx: 987.000000, Wy: 0.000000},
	"arrowup":        {GlyphName: "arrowup", Wx: 603.000000, Wy: 0.000000},
	"arrowvertex":    {GlyphName: "arrowvertex", Wx: 603.000000, Wy: 0.000000},
	"asteriskmath":   {GlyphName: "asteriskmath", Wx: 500.000000, Wy: 0.000000},
	"bar":            {GlyphName: "bar", Wx: 200.000000, Wy: 0.000000},
	"beta":           {GlyphName: "beta", Wx: 549.000000, Wy: 0.000000},
	"braceex":        {GlyphName: "braceex", Wx: 494.000000, Wy: 0.000000},
	"braceleft":      {GlyphName: "braceleft", Wx: 480.000000, Wy: 0.000000},
	"braceleftbt":    {GlyphName: "braceleftbt", Wx: 494.000000, Wy: 0.000000},
	"braceleftmid":   {GlyphName: "braceleftmid", Wx: 494.000000, Wy: 0.000000},
	"bracelefttp":    {GlyphName: "bracelefttp", Wx: 494.000000, Wy: 0.000000},
	"braceright":     {GlyphName: "braceright", Wx: 480.000000, Wy: 0.000000},
	"bracerightbt":   {GlyphName: "bracerightbt", Wx: 494.000000, Wy: 0.000000},
	"bracerightmid":  {GlyphName: "bracerightmid", Wx: 494.000000, Wy: 0.000000},
	"bracerighttp":   {GlyphName: "bracerighttp", Wx: 494.000000, Wy: 0.000000},
	"bracketleft":    {GlyphName: "bracketleft", Wx: 333.000000, Wy: 0.000000},
	"bracketleftbt":  {GlyphName: "bracketleftbt", Wx: 384.000000, Wy: 0.000000},
	"bracketleftex":  {GlyphName: "bracketleftex", Wx: 384.000000, Wy: 0.000000},
	"bracketlefttp":  {GlyphName: "bracketlefttp", Wx: 384.000000, Wy: 0.000000},
	"bracketright":   {GlyphName: "bracketright", Wx: 333.000000, Wy: 0.000000},
	"bracketrightbt": {GlyphName: "bracketrightbt", Wx: 384.000000, Wy: 0.000000},
	"bracketrightex": {GlyphName: "bracketrightex", Wx: 384.000000, Wy: 0.000000},
	"bracketrighttp": {GlyphName: "bracketrighttp", Wx: 384.000000, Wy: 0.000000},
	"bullet":         {GlyphName: "bullet", Wx: 460.000000, Wy: 0.000000},
	"carriagereturn": {GlyphName: "carriagereturn", Wx: 658.000000, Wy: 0.000000},
	"chi":            {GlyphName: "chi", Wx: 549.000000, Wy: 0.000000},
	"circlemultiply": {GlyphName: "circlemultiply", Wx: 768.000000, Wy: 0.000000},
	"circleplus":     {GlyphName: "circleplus", Wx: 768.000000, Wy: 0.000000},
	"club":           {GlyphName: "club", Wx: 753.000000, Wy: 0.000000},
	"colon":          {GlyphName: "colon", Wx: 278.000000, Wy: 0.000000},
	"comma":          {GlyphName: "comma", Wx: 250.000000, Wy: 0.000000},
	"congruent":      {GlyphName: "congruent", Wx: 549.000000, Wy: 0.000000},
	"copyrightsans":  {GlyphName: "copyrightsans", Wx: 790.000000, Wy: 0.000000},
	"copyrightserif": {GlyphName: "copyrightserif", Wx: 790.000000, Wy: 0.000000},
	"degree":         {GlyphName: "degree", Wx: 400.000000, Wy: 0.000000},
	"delta":          {GlyphName: "delta", Wx: 494.000000, Wy: 0.000000},
	"diamond":        {GlyphName: "diamond", Wx: 753.000000, Wy: 0.000000},
	"divide":         {GlyphName: "divide", Wx: 549.000000, Wy: 0.000000},
	"dotmath":        {GlyphName: "dotmath", Wx: 250.000000, Wy: 0.000000},
	"eight":          {GlyphName: "eight", Wx: 500.000000, Wy: 0.000000},
	"element":        {GlyphName: "element", Wx: 713.000000, Wy: 0.000000},
	"ellipsis":       {GlyphName: "ellipsis", Wx: 1000.000000, Wy: 0.000000},
	"emptyset":       {GlyphName: "emptyset", Wx: 823.000000, Wy: 0.000000},
	"epsilon":        {GlyphName: "epsilon", Wx: 439.000000, Wy: 0.000000},
	"equal":          {GlyphName: "equal", Wx: 549.000000, Wy: 0.000000},
	"equivalence":    {GlyphName: "equivalence", Wx: 549.000000, Wy: 0.000000},
	"eta":            {GlyphName: "eta", Wx: 603.000000, Wy: 0.000000},
	"exclam":         {GlyphName: "exclam", Wx: 333.000000, Wy: 0.000000},
	"existential":    {GlyphName: "existential", Wx: 549.000000, Wy: 0.000000},
	"five":           {GlyphName: "five", Wx: 500.000000, Wy: 0.000000},
	"florin":         {GlyphName: "florin", Wx: 500.000000, Wy: 0.000000},
	"four":           {GlyphName: "four", Wx: 500.000000, Wy: 0.000000},
	"fraction":       {GlyphName: "fraction", Wx: 167.000000, Wy: 0.000000},
	"gamma":          {GlyphName: "gamma", Wx: 411.000000, Wy: 0.000000},
	"gradient":       {GlyphName: "gradient", Wx: 713.000000, Wy: 0.000000},
	"greater":        {GlyphName: "greater", Wx: 549.000000, Wy: 0.000000},
	"greaterequal":   {GlyphName: "greaterequal", Wx: 549.000000, Wy: 0.000000},
	"heart":          {GlyphName: "heart", Wx: 753.000000, Wy: 0.000000},
	"infinity":       {GlyphName: "infinity", Wx: 713.000000, Wy: 0.000000},
	"integral":       {GlyphName: "integral", Wx: 274.000000, Wy: 0.000000},
	"integralbt":     {GlyphName: "integralbt", Wx: 686.000000, Wy: 0.000000},
	"integralex":     {GlyphName: "integralex", Wx: 686.000000, Wy: 0.000000},
	"integraltp":     {GlyphName: "integraltp", Wx: 686.000000, Wy: 0.000000},
	"intersection":   {GlyphName: "intersection", Wx: 768.000000, Wy: 0.000000},
	"iota":           {GlyphName: "iota", Wx: 329.000000, Wy: 0.000000},
	"kappa":          {GlyphName: "kappa", Wx: 549.000000, Wy: 0.000000},
	"lambda":         {GlyphName: "lambda", Wx: 549.000000, Wy: 0.000000},
	"less":           {GlyphName: "less", Wx: 549.000000, Wy: 0.000000},
	"lessequal":      {GlyphName: "lessequal", Wx: 549.000000, Wy: 0.000000},
	"logicaland":     {GlyphName: "logicaland", Wx: 603.000000, Wy: 0.000000},
	"logicalnot":     {GlyphName: "logicalnot", Wx: 713.000000, Wy: 0.000000},
	"logicalor":      {GlyphName: "logicalor", Wx: 603.000000, Wy: 0.000000},
	"lozenge":        {GlyphName: "lozenge", Wx: 494.000000, Wy: 0.000000},
	"minus":          {GlyphName: "minus", Wx: 549.000000, Wy: 0.000000},
	"minute":         {GlyphName: "minute", Wx: 247.000000, Wy: 0.000000},
	"mu":             {GlyphName: "mu", Wx: 576.000000, Wy: 0.000000},
	"multiply":       {GlyphName: "multiply", Wx: 549.000000, Wy: 0.000000},
	"nine":           {GlyphName: "nine", Wx: 500.000000, Wy: 0.000000},
	"notelement":     {GlyphName: "notelement", Wx: 713.000000, Wy: 0.000000},
	"notequal":       {GlyphName: "notequal", Wx: 549.000000, Wy: 0.000000},
	"notsubset":      {GlyphName: "notsubset", Wx: 713.000000, Wy: 0.000000},
	"nu":             {GlyphName: "nu", Wx: 521.000000, Wy: 0.000000},
	"numbersign":     {GlyphName: "numbersign", Wx: 500.000000, Wy: 0.000000},
	"omega":          {GlyphName: "omega", Wx: 686.000000, Wy: 0.000000},
	"omega1":         {GlyphName: "omega1", Wx: 713.000000, Wy: 0.000000},
	"omicron":        {GlyphName: "omicron", Wx: 549.000000, Wy: 0.000000},
	"one":            {GlyphName: "one", Wx: 500.000000, Wy: 0.000000},
	"parenleft":      {GlyphName: "parenleft", Wx: 333.000000, Wy: 0.000000},
	"parenleftbt":    {GlyphName: "parenleftbt", Wx: 384.000000, Wy: 0.000000},
	"parenleftex":    {GlyphName: "parenleftex", Wx: 384.000000, Wy: 0.000000},
	"parenlefttp":    {GlyphName: "parenlefttp", Wx: 384.000000, Wy: 0.000000},
	"parenright":     {GlyphName: "parenright", Wx: 333.000000, Wy: 0.000000},
	"parenrightbt":   {GlyphName: "parenrightbt", Wx: 384.000000, Wy: 0.000000},
	"parenrightex":   {GlyphName: "parenrightex", Wx: 384.000000, Wy: 0.000000},
	"parenrighttp":   {GlyphName: "parenrighttp", Wx: 384.000000, Wy: 0.000000},
	"partialdiff":    {GlyphName: "partialdiff", Wx: 494.000000, Wy: 0.000000},
	"percent":        {GlyphName: "percent", Wx: 833.000000, Wy: 0.000000},
	"period":         {GlyphName: "period", Wx: 250.000000, Wy: 0.000000},
	"perpendicular":  {GlyphName: "perpendicular", Wx: 658.000000, Wy: 0.000000},
	"phi":            {GlyphName: "phi", Wx: 521.000000, Wy: 0.000000},
	"phi1":           {GlyphName: "phi1", Wx: 603.000000, Wy: 0.000000},
	"pi":             {GlyphName: "pi", Wx: 549.000000, Wy: 0.000000},
	"plus":           {GlyphName: "plus", Wx: 549.000000, Wy: 0.000000},
	"plusminus":      {GlyphName: "plusminus", Wx: 549.000000, Wy: 0.000000},
	"product":        {GlyphName: "product", Wx: 823.000000, Wy: 0.000000},
	"propersubset":   {GlyphName: "propersubset", Wx: 713.000000, Wy: 0.000000},
	"propersuperset": {GlyphName: "propersuperset", Wx: 713.000000, Wy: 0.000000},
	"proportional":   {GlyphName: "proportional", Wx: 713.000000, Wy: 0.000000},
	"psi":            {GlyphName: "psi", Wx: 686.000000, Wy: 0.000000},
	"question":       {GlyphName: "question", Wx: 444.000000, Wy: 0.000000},
	"radical":        {GlyphName: "radical", Wx: 549.000000, Wy: 0.000000},
	"radicalex":      {GlyphName: "radicalex", Wx: 500.000000, Wy: 0.000000},
	"reflexsubset":   {GlyphName: "reflexsubset", Wx: 713.000000, Wy: 0.000000},
	"reflexsuperset": {GlyphName: "reflexsuperset", Wx: 713.000000, Wy: 0.000000},
	"registersans":   {GlyphName: "registersans", Wx: 790.000000, Wy: 0.000000},
	"registerserif":  {GlyphName: "registerserif", Wx: 790.000000, Wy: 0.000000},
	"rho":            {GlyphName: "rho", Wx: 549.000000, Wy: 0.000000},
	"second":         {GlyphName: "second", Wx: 411.000000, Wy: 0.000000},
	"semicolon":      {GlyphName: "semicolon", Wx: 278.000000, Wy: 0.000000},
	"seven":          {GlyphName: "seven", Wx: 500.000000, Wy: 0.000000},
	"sigma":          {GlyphName: "sigma", Wx: 603.000000, Wy: 0.000000},
	"sigma1":         {GlyphName: "sigma1", Wx: 439.000000, Wy: 0.000000},
	"similar":        {GlyphName: "similar", Wx: 549.000000, Wy: 0.000000},
	"six":            {GlyphName: "six", Wx: 500.000000, Wy: 0.000000},
	"slash":          {GlyphName: "slash", Wx: 278.000000, Wy: 0.000000},
	"space":          {GlyphName: "space", Wx: 250.000000, Wy: 0.000000},
	"spade":          {GlyphName: "spade", Wx: 753.000000, Wy: 0.000000},
	"suchthat":       {GlyphName: "suchthat", Wx: 439.000000, Wy: 0.000000},
	"summation":      {GlyphName: "summation", Wx: 713.000000, Wy: 0.000000},
	"tau":            {GlyphName: "tau", Wx: 439.000000, Wy: 0.000000},
	"therefore":      {GlyphName: "therefore", Wx: 863.000000, Wy: 0.000000},
	"theta":          {GlyphName: "theta", Wx: 521.000000, Wy: 0.000000},
	"theta1":         {GlyphName: "theta1", Wx: 631.000000, Wy: 0.000000},
	"three":          {GlyphName: "three", Wx: 500.000000, Wy: 0.000000},
	"trademarksans":  {GlyphName: "trademarksans", Wx: 786.000000, Wy: 0.000000},
	"trademarkserif": {GlyphName: "trademarkserif", Wx: 890.000000, Wy: 0.000000},
	"two":            {GlyphName: "two", Wx: 500.000000, Wy: 0.000000},
	"underscore":     {GlyphName: "underscore", Wx: 500.000000, Wy: 0.000000},
	"union":          {GlyphName: "union", Wx: 768.000000, Wy: 0.000000},
	"universal":      {GlyphName: "universal", Wx: 713.000000, Wy: 0.000000},
	"upsilon":        {GlyphName: "upsilon", Wx: 576.000000, Wy: 0.000000},
	"weierstrass":    {GlyphName: "weierstrass", Wx: 987.000000, Wy: 0.000000},
	"xi":             {GlyphName: "xi", Wx: 493.000000, Wy: 0.000000},
	"zero":           {GlyphName: "zero", Wx: 500.000000, Wy: 0.000000},
	"zeta":           {GlyphName: "zeta", Wx: 494.000000, Wy: 0.000000},
}
