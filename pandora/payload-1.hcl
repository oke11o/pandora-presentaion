request "list_req" {
  method  = "GET"
  uri = "/list?sleep=100"
  headers = {
    Authorization = "Bearer dBKAAwLnYDJGGqYyIynJIfVwZgUUsjFoSmHiOZRNEKgVXqiafXSUCqkXiZwgczfY-4"
    Content-Type  = "application/json"
    Useragent     = "Yandex"
  }
  tag = "list"
}

scenario "scenario_name" {
  requests = [
    "list_req",
  ]
}
