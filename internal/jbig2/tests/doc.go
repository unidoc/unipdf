/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

// Package tests contains the tests and benchmarks for the JBIG2 decoder and encoder.
// In order to invoke the tests the environmental variable 'UNIDOC_JBIG2_TESTDATA' with the path to
// the directory containing pdf files with some jbig2 filters must be provided.
// This package provides also a test files for the black/white image conversion as well as
// bw conversion -> encoding -> decoding process.
// These tests uses 'UNIDOC_JBIG2_TEST_IMAGES' environmental variable which should contain the path
// to the directory containing test image files.
package tests
