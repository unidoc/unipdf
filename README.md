# UniPDF - PDF for Go

[UniDoc](http://unidoc.io)'s UniPDF (formerly unidoc) is a PDF library for Go (golang) with capabilities for
creating and reading, processing PDF files. The library is written and supported by 
[FoxyUtils.com](https://foxyutils.com), where the library is used to power many of its services. 

[![Build Status](https://app.wercker.com/status/22b50db125a6d376080f3f0c80d085fa/s/master "wercker status")](https://app.wercker.com/project/bykey/22b50db125a6d376080f3f0c80d085fa)
[![GitHub (pre-)release](https://img.shields.io/github/release/unidoc/unipdf/all.svg)](https://github.com/unidoc/unipdf/releases)
[![License: AGPL v3](https://img.shields.io/badge/License-Dual%20AGPL%20v3/Commercial-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/unidoc/unipdf)](https://goreportcard.com/report/github.com/unidoc/unipdf)
[![GoDoc](https://godoc.org/github.com/unidoc/unipdf?status.svg)](https://godoc.org/github.com/unidoc/unipdf)

## Features

- [Create PDF reports](https://github.com/unidoc/unipdf-examples/blob/v3/report/pdf_report.go). Example output: [unidoc-report.pdf](https://github.com/unidoc/unipdf-examples/blob/v3/report/unidoc-report.pdf).
- [Table PDF reports](https://github.com/unidoc/unipdf-examples/blob/v3/report/pdf_tables.go). Example output: [unipdf-tables.pdf](https://github.com/unidoc/unipdf-examples/blob/v3/report/unipdf-tables.pdf).
- [Invoice creation](https://unidoc.io/news/simple-invoices)
- Paragraph in creator handling multiple styles within the same paragraph
- [Merge PDF pages](https://github.com/unidoc/unipdf-examples/blob/v3/pages/pdf_merge.go)
- [Split PDF pages](https://github.com/unidoc/unipdf-examples/blob/v3/pages/pdf_split.go) and change page order
- [Rotate pages](https://github.com/unidoc/unipdf-examples/blob/v3/pages/pdf_rotate.go)
- [Extract text from PDF files](https://github.com/unidoc/unipdf-examples/blob/v3/text/pdf_extract_text.go)
- [Text extraction support with size, position and formatting info](https://github.com/unidoc/unipdf-examples/blob/v3/text/pdf_text_locations.go)
- [PDF to CSV](https://github.com/unidoc/unipdf-examples/blob/v3/text/pdf_to_csv.go) illustrates extracting tabular data from PDF.
- [Extract images](https://github.com/unidoc/unipdf-examples/blob/v3/image/pdf_extract_images.go) with coordinates
- [Images to PDF](https://github.com/unidoc/unipdf-examples/blob/v3/image/pdf_images_to_pdf.go)
- [Add images to pages](https://github.com/unidoc/unipdf-examples/blob/v3/image/pdf_add_image_to_page.go)
- [Compress and optimize PDF](https://github.com/unidoc/unipdf-examples/blob/v3/compress/pdf_optimize.go)
- [Watermark PDF files](https://github.com/unidoc/unipdf-examples/blob/v3/image/pdf_watermark_image.go)
- Advanced page manipulation (blocks/templates)
- Load PDF templates and modify
- [Form creation](https://github.com/unidoc/unipdf-examples/blob/v3/forms/pdf_form_add.go)
- [Fill and flatten forms](https://github.com/unidoc/unipdf-examples/blob/v3/forms/pdf_form_flatten.go)
- [Fill out forms](https://github.com/unidoc/unipdf-examples/blob/v3/forms/pdf_form_fill_json.go) and [FDF merging](https://github.com/unidoc/unipdf-examples/blob/v3/forms/pdf_form_fill_fdf_merge.go)
- [Unlock PDF files / remove password](https://github.com/unidoc/unipdf-examples/blob/v3/security/pdf_unlock.go)
- [Protect PDF files with a password](https://github.com/unidoc/unipdf-examples/blob/v3/security/pdf_protect.go)
- [Digital signing validation and signing](https://github.com/unidoc/unipdf-examples/tree/v3/signatures)
- CCITTFaxDecode decoding and encoding support
- JBIG2 decoding support

Multiple examples are provided in our example repository https://github.com/unidoc/unidoc-examples.

Contact us if you need any specific examples.

## Installation
With modules:
~~~
go get github.com/unidoc/unipdf/v3
~~~

With GOPATH:
~~~
go get github.com/unidoc/unipdf/...
~~~


## How can I convince myself and my boss to buy unipdf rather using a free alternative?

The choice is yours. There are multiple respectable efforts out there that can do many good things.

In UniDoc, we work hard to provide production quality builds taking every detail into consideration and providing excellent support to our customers.  See our [testimonials](https://unidoc.io) for example.

Security.  We take security very seriously and we restrict access to github.com/unidoc/unipdf repository with protected branches and only the founders have access and every commit is reviewed prior to being accepted.

The profits are invested back into making unipdf better. We want to make the best possible product and in order to do that we need the best people to contribute. A large fraction of the profits made goes back into developing unipdf.  That way we have been able to get many excellent people to work and contribute to unipdf that would not be able to contribute their work for free.


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

Please see [pricing](https://unidoc.io/unipdf/#unipdf-pricing) to purchase a commercial license or contact sales at sales@unidoc.io
for more info.

## Getting Rid of the Watermark - Get a License
Out of the box - unipdf is unlicensed and outputs a watermark on all pages, perfect for prototyping.
To use unipdf in your projects, you need to get a license.

Get your license on [https://unidoc.io](https://unidoc.io).

The easiest way to load your license is through environment variables, for example:
```bash
export UNIPDF_CUSTOMER_NAME=UniDoc
export UNIPDF_LICENSE_PATH=/path/to/licenses/UniDoc.txt
```

Alternatively you can load the license in code, simply do:
```go
licenseKey := "... your license here ..."
customerName := `name of license holder`

err := license.SetLicenseKey(licenseKey, customerName)
if err != nil {
    fmt.Printf("Error loading license: %v\n", err)
    os.Exit(1)
}
```

[contributing]: CONTRIBUTING.md
