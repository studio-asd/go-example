#!/usr/bin/env bash

set -e

origin_dir=$PWD
schema_dir=$PWD/schemas

# vet vets all databases using SQLC vet command. The function will create a database and apply the database schema
# before invoking the 'sqlc vet' command.
#
# The command receives empty string or a valid migration directory. In the case of empty string, it will loop through
# all directories.
#
# Usage: ./all.bash vet [empty/dir_name]
vet() {
	sqlc_exec $1 vet
}

# generate generates queries from database schema/migration by invoking 'sqlc generate command'. This command will automatically
# spawn postgresql container(in local) so we have a better result when generating sqlc codes(You can look more into the issue [here](https://github.com/sqlc-dev/sqlc/issues?q=is%3Aissue+label%3Aanalyzer)).
#
# Usage: ./all.bash generate [empty/dir_name]
generate() {
	sqlc_exec $1 generate
}

# excec is an execution wrapper to run sqlc commands.
sqlc_exec() {
	pgexec="psql -U postgres"
	# Since we are not in github actions, this means we don't have the services we need to be already up and running
	# so we need to invoke docker compose in our own machine to spawn the services.
	if [[ -z "${GITHUB_ACTIONS}" ]]; then
		# Trying to bring eveyrthing down first regardless the result.
		docker compose down --remove-orphans || true
		docker compose up -d
		# Wait for the docker container to be up.
		sleep 2

		container_id=$(docker ps -aqf "name=sqlc-postgres")
		pgexec="docker exec -it ${container_id} psql -U postgres"
	fi

	# Move to schema dir.
	cd $schema_dir
	if [[ "$1" == "" ]]; then
		for dir in */; do
			# Remove all '/' suffix.
			migration_dir=${dir%/*}
			PGPASSWORD=postgres ${pgexec} -c "CREATE DATABASE ${migration_dir}"
			cd $migration_dir
			# The mounted volume is in '/data' so we need to seek the schema there.
			PGPASSWORD=postgres ${pgexec} -d ${migration_dir} -f /data/${migration_dir}/schema.sql
			sqlc $2
			if [[ "$2" = "generate" ]]; then
				go run $origin_dir/main.go gengo . --sqlc_config=sqlc.yaml --db_name=$migration_dir
			fi
			# Move to the previous directory.
			cd -
		done
	elif [[ -d ${1} ]]; then
		PGPASSWORD=postgres ${pgexec} -c "CREATE DATABASE $1"
		cd $1
		# The mounted volume is in '/data' so we need to seek the schema there.
		PGPASSWORD=postgres ${pgexec} -d $1 -f /data/$1/schema.sql
		sqlc $2
		if [[ "$2" = "generate" ]]; then
			go run $origin_dir/main.go gengo . --sqlc_config=sqlc.yaml --db_name=$migration_dir
		fi
		cd -
	fi
	# Move to before schema dir.
	cd -

	# Shutdown the services spawed by docker compose because we are not in github actions.
	if [[ -z "${GITHUB_ACTIONS}" ]]; then
		docker compose down --remove-orphans
	fi
}

# templategen creates the database directory if not exists and generate the sqlc configuration template.
# If the directory is already exists, then it will replace the sqlc coniguration in that directory with
# the new one.
#
# Usage: ./all.bash templategen [empty/dir_name]
templategen() {
	# If directory name is empty, then we should only loop through all dir and replace the template.
	if [[ "$1" == "" ]]; then
		for dir in */; do
			cp -f sqlc.yaml $dir/sqlc.yaml
			# Remove all '/' suffix.
			migration_dir=${dir%/*}
			sed -i '' "s/database_name/$migration_dir/g" $migration_dir/sqlc.yaml
		done
		exit 0;
	fi

	# Otherwise, create a new directory and generate new sqlc.yaml file.
	if [[ ! -d $1 ]]; then
		mkdir $1
	fi
	cp sqlc.yaml $1/sqlc.yaml
	sed -i '' "s/database_name/$1/g" $1/sqlc.yaml
}

case $1 in
	"vet")
		$1 $2
	;;
	"generate")
		$1 $2
	;;
	"templategen")
		$1 $2
	;;
	*)
		echo "command $1 not supported"
		exit 1
	;;
esac
