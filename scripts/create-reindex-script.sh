#!/usr/bin/env bash

# this script takes a list of instances (from a file)
# and creates the curl commands needed to asking the search API
# to reindex the dimensions for that instance
#
# along the way, it ensures that:
#
# - the state of the instance is 'published'
# - each dimension in the instance is hierarchical (checked against the hierarchy API)
#
# it creates a script (named 'adhoc-reindex.sh') which must subsequently be run
# to do the actual work of requesting reindexing
# (you may need to edit the token before running it)

###

hierarchy_api_url=localhost:10550
search_api_url=localhost:23100
search_inst_prefix=$search_api_url/search/instances
# token for accessing the search API (can easily be changed at runtime):
SERVICE_AUTH_TOKEN=changeme

# instances_json is the path to the file containing the output of 'curl http://dataset-api/instances'
instances_json=instances.json
cmd_file=adhoc-reindex.sh

####

die() { echo "$@" >&2; exit 1; }
add_cmd() { echo "$@" >> $cmd_file; }

jq_file() {
        local out=$(jq "$1" $instances_json)
        [[ -z $out ]] && die "Empty result for '$1' from $instances_json"
        echo "$out"
}

jq_str() {
        local out=$(echo "$2" | jq -r "$1")
        [[ -z $out ]] && die "Empty result for '$1' from $2"
        echo "$out"
}

curl_ok() {
        local out=$(curl -si "$1")
        if [[ $out == "HTTP/1.1 200"* ]]; then
                return 0
        elif [[ $out == "HTTP/1.1 404"* ]]; then
                return 4
        fi
        return 1
}


####


test -f $instances_json || die "No file: $instances_json"

echo Creating: $cmd_file ...
{
    cat <<-EOFstarter
	#!/usr/bin/env bash
	# XXX do not commit me XXX
	
	# change these?
	auth_token=$SERVICE_AUTH_TOKEN
	search_inst_prefix=$search_inst_prefix
	
	cyrl(){ echo "Doing \$1"; curl -iH "Authorization: Bearer \$auth_token" -X PUT "\$1" || sleep 10; }
	
EOFstarter
} > $cmd_file
chmod 750 $cmd_file

instance_count=$(jq_file '.items | length')

instances_processed=0
dimensions_count=0
dimensions_processed=0
skipped=0
errors=0
for (( idx=0 ; idx < instance_count ; idx++ )); do
        echo "Processing $(( idx+1 )) of $instance_count"

        inst_json=$(jq_file ".items[$idx]")
        state=$(jq_str ".state" "$inst_json")
        if [[ $state != published ]]; then echo "	Want 'published' got: '$state'"; continue; fi

        inst_id=$(jq_str ".id" "$inst_json")
        dims=$(jq_str ".dimensions" "$inst_json")
        dim_count=$(jq_str ".dimensions | length" "$inst_json")
        echo "	dimensions: $dim_count"

        for (( d=0 ; d < dim_count; d++ )); do
                dim=$(jq_str ".[$d].name" "$dims")
                let dimensions_count++

                # check dimension is hierarchy
                curl_res=0
                curl_ok "$hierarchy_api_url/hierarchies/$inst_id/$dim" || curl_res=$?
                if [[ $curl_res -eq 0 ]]; then
                        echo "		add hierarchy $inst_id $dim"
                        add_cmd "cyrl \"\$search_inst_prefix/$inst_id/dimensions/$dim\""
                        let dimensions_processed++
                        [[ $dimensions_processed -lt 5 ]] && add_cmd sleep $(( 6 - dimensions_processed ))
                elif [[ $curl_res -eq 4 ]]; then
                        echo "			skip non-hierarchy $inst_id $dim"
                        let skipped++
                else
                        echo "				ERROR		$inst_id $dim	***********************************************"
                        let errors++
                        [[ $errors -gt 10 ]] && die Too many errors
                fi
        done
        let instances_processed++
done

echo "Done. instances:  $instance_count inst_processed: $instances_processed"
echo "      dimensions: $dimensions_count  dim_processed: $dimensions_processed skipped: $skipped errors: $errors"
echo "Created script: $cmd_file"
[[ $errors -eq 0 ]]
