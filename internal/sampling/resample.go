/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package sampling

// ResampleBytes resamples the raw data which is in 8-bit (byte) format as a different
// bit count per sample, up to 32 bits (uint32).
func ResampleBytes(data []byte, bitsPerSample int) []uint32 {
	var samples []uint32

	bitsLeftPerSample := bitsPerSample
	var sample uint32
	var remainder byte
	remainderBits := 0

	index := 0

	i := 0
	for i < len(data) {
		// Start with the remainder.
		if remainderBits > 0 {
			take := remainderBits
			if bitsLeftPerSample < take {
				take = bitsLeftPerSample
			}

			sample = (sample << uint(take)) | uint32(remainder>>uint(8-take))
			remainderBits -= take
			if remainderBits > 0 {
				remainder = remainder << uint(take)
			} else {
				remainder = 0
			}
			bitsLeftPerSample -= take

			if bitsLeftPerSample == 0 {
				//samples[index] = sample
				samples = append(samples, sample)
				bitsLeftPerSample = bitsPerSample
				sample = 0
				index++
			}
		} else {
			// Take next byte
			b := data[i]
			i++

			// 8 bits.
			take := 8
			if bitsLeftPerSample < take {
				take = bitsLeftPerSample
			}
			remainderBits = 8 - take
			sample = (sample << uint(take)) | uint32(b>>uint(remainderBits))

			if take < 8 {
				remainder = b << uint(take)
			}

			bitsLeftPerSample -= take
			if bitsLeftPerSample == 0 {
				//samples[index] = sample
				samples = append(samples, sample)
				bitsLeftPerSample = bitsPerSample
				sample = 0
				index++
			}
		}
	}

	// Take care of remaining samples (if enough data available).
	for remainderBits >= bitsPerSample {
		take := remainderBits
		if bitsLeftPerSample < take {
			take = bitsLeftPerSample
		}

		sample = (sample << uint(take)) | uint32(remainder>>uint(8-take))
		remainderBits -= take
		if remainderBits > 0 {
			remainder = remainder << uint(take)
		} else {
			remainder = 0
		}
		bitsLeftPerSample -= take
		if bitsLeftPerSample == 0 {
			//samples[index] = sample
			samples = append(samples, sample)
			bitsLeftPerSample = bitsPerSample
			sample = 0
			index++
		}
	}

	return samples
}

// ResampleUint32 resamples the raw data which is in <=32-bit (uint32) format as a different
// bit count per sample, up to 32 bits (uint32).
//
// bitsPerOutputSample is the number of bits for each output sample (up to 32)
// bitsPerInputSample is the number of bits used in each input sample (up to 32)
func ResampleUint32(data []uint32, bitsPerInputSample int, bitsPerOutputSample int) []uint32 {
	var samples []uint32

	bitsLeftPerSample := bitsPerOutputSample
	var sample uint32
	var remainder uint32
	remainderBits := 0

	index := 0

	i := 0
	for i < len(data) {
		// Start with the remainder.
		if remainderBits > 0 {
			take := remainderBits
			if bitsLeftPerSample < take {
				take = bitsLeftPerSample
			}

			sample = (sample << uint(take)) | uint32(remainder>>uint(bitsPerInputSample-take))
			remainderBits -= take
			if remainderBits > 0 {
				remainder = remainder << uint(take)
			} else {
				remainder = 0
			}
			bitsLeftPerSample -= take
			if bitsLeftPerSample == 0 {
				//samples[index] = sample
				samples = append(samples, sample)
				bitsLeftPerSample = bitsPerOutputSample
				sample = 0
				index++
			}
		} else {
			// Take next byte
			b := data[i]
			i++

			// 32 bits.
			take := bitsPerInputSample
			if bitsLeftPerSample < take {
				take = bitsLeftPerSample
			}
			remainderBits = bitsPerInputSample - take
			sample = (sample << uint(take)) | uint32(b>>uint(remainderBits))

			if take < bitsPerInputSample {
				remainder = b << uint(take)
			}

			bitsLeftPerSample -= take
			if bitsLeftPerSample == 0 {
				//samples[index] = sample
				samples = append(samples, sample)
				bitsLeftPerSample = bitsPerOutputSample
				sample = 0
				index++
			}
		}
	}

	// Take care of remaining samples (if enough data available).
	for remainderBits >= bitsPerOutputSample {
		take := remainderBits
		if bitsLeftPerSample < take {
			take = bitsLeftPerSample
		}

		sample = (sample << uint(take)) | uint32(remainder>>uint(bitsPerInputSample-take))
		remainderBits -= take
		if remainderBits > 0 {
			remainder = remainder << uint(take)
		} else {
			remainder = 0
		}
		bitsLeftPerSample -= take
		if bitsLeftPerSample == 0 {
			samples = append(samples, sample)
			bitsLeftPerSample = bitsPerOutputSample
			sample = 0
			index++
		}
	}

	// If there are partial output samples, pad with 0s.
	if bitsLeftPerSample > 0 && bitsLeftPerSample < bitsPerOutputSample {
		sample <<= uint(bitsLeftPerSample)
		samples = append(samples, sample)
	}

	return samples
}
