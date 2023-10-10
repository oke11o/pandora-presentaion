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
      first_item_id = "$.items[0]"
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
{"item_id": {{.request.list_req.postprocessor.first_item_id}}}
EOF
  tag  = "order_req"
}

scenario "scenario_name" {
  requests = [
    "list_req",
    "order_req"
  ]
}
