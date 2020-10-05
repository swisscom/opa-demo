package swisscom.example

# HTTP API request
import input

# bob is alice's manager, and betty is charlie's.
subordinates = {"alice": [], "charlie": [], "bob": ["alice"], "betty": ["charlie"]}


default allow = false

# Allow users to get their profiles
allow {
  some username
  input.method == "GET"
  input.path = ["api", "v1", "employees", username]
  input.user == username
}

# Allow managers to get their subordinates profiles
allow {
  some username
  input.method == "GET"
  input.path = ["api", "v1", "employees", username]
  subordinates[input.user][_] == username
}
