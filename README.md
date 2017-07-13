# Version 2

The version 2 of UniDoc is currently in alpha. The pdf functionality has been split into two parts.  The core subpackage contains core PDF file parsing functionality and primitive objects, whereas the model subpackage provides a higher level interface to the PDF.

# Migrating from version 1.

Migrating is fairly straightforward.  A few things are incompatible and will be listed here prior to release.

---

# UniDoc

[UniDoc](http://unidoc.io) is a fast and powerful open source library for document manipulation starting off as a PDF
toolkit. This is a library written and supported by the owners
of the [FoxyUtils.com](https://foxyutils.com) website.

This library is used to power many of the PDF services offered by [FoxyUtils](https://foxyutils.com). The goal is to extend it to
eventually support all of the offered services.

[![wercker status](https://app.wercker.com/status/22b50db125a6d376080f3f0c80d085fa/s/master "wercker status")](https://app.wercker.com/project/bykey/22b50db125a6d376080f3f0c80d085fa)
[![GoDoc](https://godoc.org/github.com/unidoc/unidoc?status.svg)](https://godoc.org/github.com/unidoc/unidoc)

## Installation
~~~
go get github.com/unidoc/unidoc
~~~

## Vendoring
For reliability, we recommend using specific versions and the vendoring capability of golang.
Check out the Releases section to see the tagged releases.

## Overview

 * Many [features](http://unidoc.io/features) with documented examples.
 * Self contained with no external dependencies
 * Developer friendly

## Examples

See the [unidoc-examples](https://github.com/unidoc/unidoc-examples/tree/master) folder.

## Copying/License

UniDoc is licensed as [AGPL][agpl] software (with extra terms as specified in our license).

AGPL is a free / open source software license.

This doesn't mean the software is gratis!

Buying a license is mandatory as soon as you develop activities
distributing the UniDoc software inside your product or deploying it on a network
without disclosing the source code of your own applications under the AGPL license.
These activities include:

 * offering services as an application service provider or over-network application programming interface (API)
 * creating/manipulating documents for users in a web/server/cloud application
 * shipping UniDoc with a closed source product

Contact sales for more info: sales@unidoc.io.

## Contributing

Contributors need to approve the [Contributor License Agreement](https://docs.google.com/a/owlglobal.io/forms/d/1PfTjEAi67-x0JOTU45SDonJnWy1fWB_J1aopGss34bY/viewform) before any code will be reviewed. Preferably add a test case to make sure there is no regression and that the new behaviour is as expected.

## Support

Please email us at support@unidoc.io for any queries.

## Stay up to date

* Follow us on [twitter](https://twitter.com/unidoclib)
* Sign-up for our [newsletter](http://eepurl.com/b9Idt9)

[agpl]: LICENSE.md
[contributing]: CONTRIBUTING.md
