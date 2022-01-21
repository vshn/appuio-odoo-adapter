# Odoo Adapter for APPUiO Cloud

[![Build](https://img.shields.io/github/workflow/status/vshn/appuio-odoo-adapter/Test)][build]
![Go version](https://img.shields.io/github/go-mod/go-version/vshn/appuio-odoo-adapter)
[![Version](https://img.shields.io/github/v/release/vshn/appuio-odoo-adapter)][releases]
[![Maintainability](https://img.shields.io/codeclimate/maintainability/vshn/appuio-odoo-adapter)][codeclimate]
[![Coverage](https://img.shields.io/codeclimate/coverage/vshn/appuio-odoo-adapter)][codeclimate]
[![GitHub downloads](https://img.shields.io/github/downloads/vshn/appuio-odoo-adapter/total)][releases]

[build]: https://github.com/vshn/appuio-odoo-adapter/actions?query=workflow%3ATest
[releases]: https://github.com/vshn/appuio-odoo-adapter/releases
[codeclimate]: https://codeclimate.com/github/vshn/appuio-odoo-adapter

[APPUiO Cloud](https://appuio.cloud) is based on OpenShift 4 and can be considered a "Namespace as a Service".
It follows a pay-per-use pricing model, however APPUiO Cloud does not include or require a specific accounting or billing software (ERP).

Instead, the ERP software is pluggable and APPUiO Cloud specifies the "interface".
Such a plugin is called "adapter" in this context.

VSHN uses currently Odoo as its ERP software and this repository "implements" the adapter to serve as a bridge between APPUiO Cloud and Odoo.

**Architecture documentation**: https://kb.vshn.ch/appuio-cloud

**Adapter documentation**: https://vshn.github.io/appuio-odoo-adapter
