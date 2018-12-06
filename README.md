# Version 3 - Upcoming

Version 3 of UniDoc is currently in alpha. It marks multiple significant new major advancements as well as many fixes and enhancements:

- [ ] Composite fonts supported and font handling has and is being completely revamped, including unicode support.
- [ ] Digital signing validation and signing
- [ ] Append mode
- [x] PDF compression and optimization of outputs with several options 1) combining duplicates, 2) compressed object streams, 3) image points per inch threshold, 4) image quality.
- [ ] Text extraction significantly improved in quality and support for vectorized (position-based) text extraction (XY)
- [x] Paragraph in creator handling multiple styles within the same paragraph
- [x] Invoice component for easy PDF invoice generation
- [x] Table of contents automatically generated
- [x] Encryption support refactored and AESv3 support added
- [x] Form field filling and form flattening with appearance generation
- [x] Getting form field values and listing
- [x] FDF merge

Go give it a spin, checkout the `v3` branch of unidoc and `v3` branch of unidoc-examples:
- https://github.com/unidoc/unidoc/tree/v3
- https://github.com/unidoc/unidoc-examples/tree/v3

---

# UniDoc

[UniDoc](http://unidoc.io) is a powerful PDF library for Go (golang). The library is written and supported by the owners of the [FoxyUtils.com](https://foxyutils.com) website, where the library is used to power many of the PDF services offered. 

[![wercker status](https://app.wercker.com/status/22b50db125a6d376080f3f0c80d085fa/s/master "wercker status")](https://app.wercker.com/project/bykey/22b50db125a6d376080f3f0c80d085fa)
[![GoDoc](https://godoc.org/github.com/unidoc/unidoc?status.svg)](https://godoc.org/github.com/unidoc/unidoc)

## Installation
~~~
go get github.com/unidoc/unidoc/...
~~~

## Features
UniDoc has a powerful set of features both for reading, processing and writing PDF.
The following list describes some of the main features:

- [x] Create PDF reports with easy interface
- [x] Create PDF invoices (v3)
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
- [ ] Digital signatures (v3)


## How can I convince myself and my boss to buy unidoc rather using a free alternative?

The choice is yours. There are multiple respectable efforts out there that can do many good things.

In unidoc, we work hard to provide production quality builds taking every detail into consideration and providing excellent support to our customers.  See our [testimonials](https://unidoc.io) for example.

Security.  We take security very seriously and we restrict access to github.com/unidoc/unidoc repository with protected branches and only 2 of the founders have access and every commit is reviewed prior to being accepted.

The money goes back into making unidoc better. We want to make the best possible product and in order to do that we need the best people to contribute. A large fraction of the profits made goes back into developing unidoc.  That way we have been able to get many excellent people to work and contribute to unidoc that would not be able to contribute their work for free.


## Examples

Multiple examples are provided in our example repository.
Many features for processing PDF files with [documented examples](https://unidoc.io/examples) on our website.

Contact us if you need any specific examples.

## Vendoring
For reliability, we recommend using specific versions and the vendoring capability of golang.
Check out the Releases section to see the tagged releases.


## Contributing

Contributors need to approve the [Contributor License Agreement](https://docs.google.com/a/owlglobal.io/forms/d/1PfTjEAi67-x0JOTU45SDonJnWy1fWB_J1aopGss34bY/viewform) before any code will be reviewed. Preferably add a test case to make sure there is no regression and that the new behaviour is as expected.

## Support and consulting

Please email us at support@unidoc.io for any queries.

If you have any specific tasks that need to be done, we offer consulting in certain cases.
Please contact us with a brief summary of what you need and we will get back to you with a quote, if appropriate.

## Licensing Information

This library (UniDoc) has a dual license, a commercial one suitable for closed source projects and an
AGPL license that can be used in open source software.

Depending on your needs, you must choose one of them and follow its policies. A detail of the policies
and agreements for each license type are available in the [LICENSE.COMMERCIAL](LICENSE.COMMERCIAL)
and [LICENSE.AGPL](LICENSE.AGPL) files.

In brief, purchasing a license is mandatory as soon as you develop activities
distributing the UniDoc software inside your product or deploying it on a network
without disclosing the source code of your own applications under the AGPL license.
These activities include:

 * offering services as an application service provider or over-network application programming interface (API)
 * creating/manipulating documents for users in a web/server/cloud application
 * shipping UniDoc with a closed source product

Please see [pricing](http://unidoc.io/pricing) to purchase a commercial license or contact sales at sales@unidoc.io
for more info.

## Getting Rid of the Watermark - Get a License
Out of the box - unidoc is unlicensed and outputs a watermark on all pages, perfect for prototyping.
To use unidoc in your projects, you need to get a license.

Get your license on [https://unidoc.io](https://unidoc.io).

To load your license, simply do:
```
unidocLicenseKey := "... your license here ..."
err := license.SetLicenseKey(unidocLicenseKey)
if err != nil {
    fmt.Printf("Error loading license: %v\n", err)
    os.Exit(1)
}
```


[contributing]: CONTRIBUTING.md
