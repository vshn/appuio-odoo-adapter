# Odoo Adapter for APPUiO Cloud

Test closing.

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

## Quick Start

### Generate Invoices

```sh
kubectl -n appuio-reporting port-forward svc/reporting-db 5432 &
kubectl -n vshn-odoo-test port-forward svc/odoo 8080:8000 &

DB_USER=$(kubectl -n appuio-reporting get secret/reporting-db-superuser -o jsonpath='{.data.user}' | base64 --decode)
DB_PASSWORD=$(kubectl -n appuio-reporting get secret/reporting-db-superuser -o jsonpath='{.data.password}' | base64 --decode)

export OA_DB_URL="postgres://${DB_USER}:${DB_PASSWORD}@localhost/reporting?sslmode=disable"
export OA_ODOO_URL="http://admin:${ODOO_PASSWORD}@localhost:8080/VSHNProd_2022-01-31"

go run . invoice --year 2022 --month 1
```

## Documentation

**Architecture documentation**: https://kb.vshn.ch/appuio-cloud

**Adapter documentation**: https://vshn.github.io/appuio-odoo-adapter
