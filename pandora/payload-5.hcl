variable_source "users" "file/csv" {
  file              = "users.csv"
  fields            = ["user_id", "name", "pass"]
  ignore_first_line = true
  delimiter         = ","
}

request "auth_req" {
  method  = "POST"
  uri     = "/auth"
  headers = {
    Content-Type = "application/json"
    Useragent    = "Yandex"
  }
  body = <<EOF
{"user_id":  {{.request.auth_req.preprocessor.user_id}}}
EOF
  tag  = "auth"

  preprocessor {
    mapping = {
      user_id = "source.users[next].user_id"
    }
  }
  postprocessor "var/jsonpath" {
    mapping = {
      token = "$.auth_key"
    }
  }
}

request "list_req" {
  method  = "GET"
  uri     = "/list?sleep=100"
  headers = {
    Authorization = "Bearer {{.request.auth_req.postprocessor.token}}"
    Content-Type  = "application/json"
    Useragent     = "Yandex"
  }
  tag = "list"

  postprocessor "var/jsonpath" {
    mapping = {
      items = "$.items"
    }
  }
}

request "order_req" {
  method  = "POST"
  uri     = "/order?sleep=100"
  headers = {
    Authorization = "Bearer {{.request.auth_req.postprocessor.token}}"
    Content-Type  = "application/json"
    Useragent     = "Yandex"
  }
  body = <<EOF
{"item_id": {{.request.order_req.preprocessor.item}}}
EOF
  tag  = "order"

  preprocessor {
    mapping = {
      item = "request.list_req.postprocessor.items[next]"
    }
  }
}

scenario "scenario_name" {
  min_waiting_time = 1000
  requests = [
    "auth_req",
    "sleep(100)",
    "list_req",
    "sleep(100)",
    "order_req(3, 100)"
  ]
}
