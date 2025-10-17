# Idea behind 3 different logging files is 3 different types of logs

## func info (file, context)
-> 2025/08/26 00:27:53 [ADMIN-INFO] {file} - {context}

## func error (file, prio, context, err)
possible prios: fatal, critical, warn
-> 2025/08/26 00:25:42 [ADMIN-ERROR-ALERT] high | {file} - {context} - {err}