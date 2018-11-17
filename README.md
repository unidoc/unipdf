# UniDoc

[UniDoc](http://unidoc.io) is a powerful PDF library for Go (golang). The library is written and supported by the owners of the [FoxyUtils.com](https://foxyutils.com) website, where the library is used to power many of the PDF services offered. 

[![wercker status](https://app.wercker.com/status/22b50db125a6d376080f3f0c80d085fa/s/master "wercker status")](https://app.wercker.com/project/bykey/22b50db125a6d376080f3f0c80d085fa)
[![GoDoc](https://godoc.org/github.com/unidoc/unidoc?status.svg)](https://godoc.org/github.com/unidoc/unidoc)

## Installation
~~~
go get github.com/unidoc/unidoc/...
~~~

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

## Examples

Multiple examples are provided in our example repository.
Many features for processing PDF files with [documented examples](https://unidoc.io/examples) on our website.

Contact us if you need any specific examples.

## Vendoring
For reliability, we recommend using specific versions and the vendoring capability of golang.
Check out the Releases section to see the tagged releases.

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

## Contributing

Contributors need to approve the [Contributor License Agreement](https://docs.google.com/a/owlglobal.io/forms/d/1PfTjEAi67-x0JOTU45SDonJnWy1fWB_J1aopGss34bY/viewform) before any code will be reviewed. Preferably add a test case to make sure there is no regression and that the new behaviour is as expected.

## Support and consulting

Please email us at support@unidoc.io for any queries.

If you have any specific tasks that need to be done, we offer consulting in certain cases.
Please contact us with a brief summary of what you need and we will get back to you with a quote, if appropriate.


[contributing]: CONTRIBUTING.md
