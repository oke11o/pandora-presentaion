
variable_source "users" "file/csv" {
  file              = "../users.csv"
  fields            = ["user_id", "name", "pass"]
  ignore_first_line = true
  delimiter         = ","
}
variable_source "filter_src" "file/json" {
  file = "../filter.json"
}
request "auth_req" {
  method = "POST"
  uri    = "/auth"
  headers = {
    Content-Type = "application/json"
    Useragent    = "Tank"
  }
  tag       = "auth"
  body      = <<EOF
{"user_id":  {{.request.auth_req.preprocessor.user_id}}}
EOF
  templater = "text"

  preprocessor {
    mapping = {
      user_id = "source.users[next].user_id"
    }
  }
  postprocessor "var/header" {
    mapping = {
      Content-Type      = "Content-Type|upper"
      httpAuthorization = "Http-Authorization"
    }
  }
  postprocessor "var/jsonpath" {
    mapping = {
      token = "$.auth_key"
    }
  }
  postprocessor "assert/response" {
    headers = {
      Content-Type = "json"
    }
    body = ["key"]
    size {
      val = 40
      op  = ">"
    }
  }
  postprocessor "assert/response" {
    body = ["auth"]
  }
}
request "list_req" {
  method = "GET"
  headers = {
    Authorization = "Bearer {{.request.auth_req.postprocessor.token}}"
    Content-Type  = "application/json"
    Useragent     = "Tank"
  }
  tag = "list"
  uri = "/list"

  postprocessor "var/jsonpath" {
    mapping = {
      item_id = "$.items[0]"
      items   = "$.items"
    }
  }
}
request "order_req" {
  method = "POST"
  uri    = "/order"
  headers = {
    Authorization = "Bearer {{.request.auth_req.postprocessor.token}}"
    Content-Type  = "application/json"
    Useragent     = "Tank"
  }
  tag  = "order_req"
  body = <<EOF
{"item_id": {{.request.order_req.preprocessor.item}}}
EOF

  preprocessor {
    mapping = {
      item = "request.list_req.postprocessor.items[next]"
    }
  }
}

request "order_req2" {
  method = "POST"
  uri    = "/order"
  headers = {
    Authorization = "Bearer {{.request.auth_req.postprocessor.token}}"
    Content-Type  = "application/json"
    Useragent     = "Tank"
  }
  tag  = "order_req"
  body = <<EOF
{"item_id": {{.request.order_req2.preprocessor.item}}  }
EOF

  preprocessor {
    mapping = {
      item = "request.list_req.postprocessor.items[next]"
    }
  }
}

scenario "scenario_name" {
  weight           = 50
  min_waiting_time = 10
  shoot = [
    "auth_req(1)",
    "sleep(100)",
    "list_req(1)",
    "sleep(100)",
    "order_req(3)"
  ]
}
