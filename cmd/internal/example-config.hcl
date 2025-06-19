schema = "$gitlab_path.strip_last_prefix(\"helm-\")/$[technologies]/+"

input "gitlab_path" "environment"{
    //type = "envvar"
    from = "GITLAB_PATH"
}

input "technologies" "variable_set"{
    content = [
        "mssql",
        "kafka",
        "postgres",
    ]
}