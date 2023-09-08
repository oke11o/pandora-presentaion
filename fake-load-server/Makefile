help:
	@echo "make auth user=6\n\
make list bearer=JscGSLzfUUmXzCEbjBARFrOgSdBNRfKuPKwWzngcuFPPzGDLiiYyLoDzCSgXImVo\n\
make item item=6000 bearer=JscGSLzfUUmXzCEbjBARFrOgSdBNRfKuPKwWzngcuFPPzGDLiiYyLoDzCSgXImVo"

# make auth user=6
auth:
	curl -X POST --location "http://127.0.0.1:8092/auth" \
        -H "Content-Type: application/json" \
        -d "{\"user_id\": $(user)}"

# make list bearer=JscGSLzfUUmXzCEbjBARFrOgSdBNRfKuPKwWzngcuFPPzGDLiiYyLoDzCSgXImVo
list:
	curl -X GET --location "http://127.0.0.1:8092/list" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer ${bearer}"

# make item item=6000 bearer=JscGSLzfUUmXzCEbjBARFrOgSdBNRfKuPKwWzngcuFPPzGDLiiYyLoDzCSgXImVo
item:
	curl -X POST --location "http://127.0.0.1:8092/item" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer ${bearer}" \
        -d "{\"item_id\": ${item}}"

stats:
	curl -X GET --location "http://127.0.0.1:8092/statistic" -H "Content-Type: application/json" | jq

reset:
	curl -X POST --location "http://127.0.0.1:8092/reset" -H "Content-Type: application/json"
