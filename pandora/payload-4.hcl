request "list_req" {
  method  = "GET"
  uri = "/list?sleep=100"
  headers = {
    Authorization = "Bearer dBKAAwLnYDJGGqYyIynJIfVwZgUUsjFoSmHiOZRNEKgVXqiafXSUCqkXiZwgczfY-4"
    Content-Type  = "application/json"
    Useragent     = "Yandex"
  }
  tag = "list"

  postprocessor "var/jsonpath" {
    mapping = {
      items   = "$.items"
    }
  }
}

request "order_req" {
  method = "POST"
  uri    = "/order?sleep=100"
  headers = {
    Authorization = "Bearer dBKAAwLnYDJGGqYyIynJIfVwZgUUsjFoSmHiOZRNEKgVXqiafXSUCqkXiZwgczfY-4"
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
  requests = [
    "list_req",
    "sleep(100)",
    "order_req(3, 100)"
  ]
}
