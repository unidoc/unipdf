# UniPDF - PDF for Go

[UniDoc](http://unidoc.io)'s unipdf is a powerful PDF library for Go (golang) with capabilities for
creating and processing PDF files. The library is written and supported by 
the [FoxyUtils.com](https://foxyutils.com) website, where the library is used to power
many of the PDF services offered. 

[![Build Status](https://app.wercker.com/status/22b50db125a6d376080f3f0c80d085fa/s/master "wercker status")](https://app.wercker.com/project/bykey/22b50db125a6d376080f3f0c80d085fa)
[![License: AGPL v3](https://img.shields.io/badge/License-Dual%20AGPL%20v3/Commercial-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/unidoc/unipdf)](https://goreportcard.com/report/github.com/unidoc/unipdf)
[![GoDoc](https://godoc.org/github.com/unidoc/unipdf?status.svg)](https://godoc.org/github.com/unidoc/unipdf)

## News
- unidoc is being renamed to unipdf and will be maintained under https://github.com/unidoc/unipdf
- The old repository will remain under https://github.com/unidoc/unipdf for backwards compatibility and will be read-only.
All development will be under the unipdf repository.
- The initial release of unipdf v3.0.0 will be compliant with Go modules from the start.


## Features
unipdf has a powerful set of features both for reading, processing and writing PDF.
The following list describes some key features:

- [x] Create PDF reports with easy interface
- [x] [Invoice component for easy PDF invoice generation](https://unidoc.io/news/simple-invoices)
- [x] Advanced table generation in the creator with subtable support
- [x] Composite fonts supported and font handling has and is being completely revamped, including unicode support.
- [x] Paragraph in creator handling multiple styles within the same paragraph
- [x] Table of contents automatically generated
- [x] PDF compression and optimization of outputs with several options 1) combining duplicates, 2) compressed object streams, 3) image points per inch threshold, 4) image quality.
- [x] Text extraction significantly improved in quality and foundation in place for vectorized (position-based) text extraction (XY)
- [x] Image extraction with coordinates
- [x] Merge PDF pages
- [x] Merge page contents
- [x] Split PDF pages and change page order
- [x] Rotate pages
- [x] Extract text from PDF files
- [x] Extract images
- [x] Add images to pages
- [x] Compress and optimize PDF output (v3)
- [x] Draw watermark on PDF files
- [x] Advanced page manipulation (blocks/templates)
- [x] Load PDF templates and modify
- [x] Flatten forms and generate appearance streams (v3)
- [x] Fill out forms and FDF merging (v3)
- [x] Unlock PDF files / remove password
- [x] Protect PDF files with a password
- [x] Digital signing validation and signing
- [x] Encryption support refactored and AESv3 support added
- [x] CCITTFaxDecode decoding and encoding support
- [x] FDF merge and form filling via JSON data
- [x] Form field filling and form flattening with appearance generation
- [x] Append mode
- [x] Getting form field values and listing

## Installation
With modules:
~~~
go get github.com/unidoc/unipdf/v3
~~~


## How can I convince myself and my boss to buy unipdf rather using a free alternative?

The choice is yours. There are multiple respectable efforts out there that can do many good things.

In UniDoc, we work hard to provide production quality builds taking every detail into consideration and providing excellent support to our customers.  See our [testimonials](https://unidoc.io) for example.

Security.  We take security very seriously and we restrict access to github.com/unidoc/unipdf repository with protected branches and only the founders have access and every commit is reviewed prior to being accepted.

The profits are invested back into making unipdf better. We want to make the best possible product and in order to do that we need the best people to contribute. A large fraction of the profits made goes back into developing unipdf.  That way we have been able to get many excellent people to work and contribute to unipdf that would not be able to contribute their work for free.


## Examples

Multiple examples are provided in our example repository https://github.com/unidoc/unidoc-examples
as well as [documented examples](https://unidoc.io/examples) on our website.

Contact us if you need any specific examples.

## Contributing

[![CLA assistant](https://cla-assistant.io/readme/badge/unidoc/unipdf)](https://cla-assistant.io/unidoc/unipdf)

All contributors must sign a contributor license agreement before their code will be reviewed and merged.

## Support and consulting

Please email us at support@unidoc.io for any queries.

If you have any specific tasks that need to be done, we offer consulting in certain cases.
Please contact us with a brief summary of what you need and we will get back to you with a quote, if appropriate.

## Licensing Information

This library (unipdf) has a dual license, a commercial one suitable for closed source projects and an
AGPL license that can be used in open source software.

Depending on your needs, you must choose one of them and follow its policies. A detail of the policies
and agreements for each license type are available in the [LICENSE.COMMERCIAL](LICENSE.COMMERCIAL)
and [LICENSE.AGPL](LICENSE.AGPL) files.

In brief, purchasing a license is mandatory as soon as you develop activities
distributing the unipdf software inside your product or deploying it on a network
without disclosing the source code of your own applications under the AGPL license.
These activities include:

 * offering services as an application service provider or over-network application programming interface (API)
 * creating/manipulating documents for users in a web/server/cloud application
 * shipping unipdf with a closed source product

Please see [pricing](http://unidoc.io/pricing) to purchase a commercial license or contact sales at sales@unidoc.io
for more info.

## Getting Rid of the Watermark - Get a License
Out of the box - unipdf is unlicensed and outputs a watermark on all pages, perfect for prototyping.
To use unipdf in your projects, you need to get a license.

Get your license on [https://unidoc.io](https://unidoc.io).

To load your license, simply do:
```go
unidocLicenseKey := "... your license here ..."
err := license.SetLicenseKey(unidocLicenseKey)
if err != nil {
    fmt.Printf("Error loading license: %v\n", err)
    os.Exit(1)
}
```

[contributing]: CONTRIBUTING.md
