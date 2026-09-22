# The spend alert. It is an alert and not a cap, because Google has no switch
# that stops a project at a number: what it buys is finding out on the day
# something started costing money, rather than at the end of the month.

data "google_project" "this" {
  project_id = var.project_id
}

resource "google_billing_budget" "spend" {
  billing_account = var.billing_account
  display_name    = "${var.service_name} spend"

  # Only this project's charges, and only what is actually charged: the free
  # allowance arrives as a credit and is subtracted before this is measured.
  budget_filter {
    projects        = ["projects/${data.google_project.this.number}"]
    calendar_period = "MONTH"
  }

  amount {
    specified_amount {
      currency_code = var.budget_currency
      units         = var.budget_amount
    }
  }

  # Half of it, all of it, and a forecast that the month will end over it.
  threshold_rules {
    threshold_percent = 0.5
  }

  threshold_rules {
    threshold_percent = 1.0
  }

  threshold_rules {
    threshold_percent = 1.0
    spend_basis       = "FORECASTED_SPEND"
  }

  depends_on = [google_project_service.enabled]
}
