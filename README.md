Heimdall
===

Heimdall is a authentication & authorization wrapper for web apps that would 
like to implement security. Heimdall implements basic http authentication, 
oauth2 bearer authentication, and cookie authentication.

Getting Started
---

To start using heimdall you need to pick how you'd like to store your users, 
clients, and tokens. There are two out of the box implementations:

- In Memory
- Filesystem

Using the in memory implementation:

	import "github.com/murphysean/heimdall"
	import "github.com/murphysean/heimdall/memdb

	hh := heimdall.NewHeimdall(http.DefaultServerMux, scopeFunc, pdpFunc, failFunc)
        hh.DB = memdb.NewMemDB()
	//Put a user into the database
	user := hh.DB.CreateUser()
	user.SetId("1")
	user.SetName("User One")
	//A hack to allow this user to authenticate
	user.(memdb.User)["username"] = "user1"
	user.(memdb.User)["password"] = "password"

	http.HandleFunc("/login", hh.Login)
	http.ListenAndServe(":http", hh)

Using the filesystem as a db requires a little setup, first you will have to 
get the directory structure set up:

	mkdir db
	cd db
	mkdir users clients tokens
	echo "1,user1,password" > login.csv
	echo '{"id":"1","name":"User One"}' > users/1.json

And now you can use the filesystem version:

	import "github.com/murphysean/heimdall"
	import "github.com/murphysean/heimdall/filedb"
	
	db := filedb.NewFileDB()
	hh := heimdall.NewHeimdall(http.DefaultServerMux, scopeFunc, pdpFunc, failFunc)
        hh.DB = filedb.NewFileDB("db")
	
	http.HandleFunc("/login", hh.Login)
	http.ListenAndServe(":http", hh)

How Heimdall Works
---

Heimdall works by seperating your handlers from authentication and 
authorization work. It will reach out into special callback functions with 
information it's pulled from the request and the heimdall database 
implementation. You will then be able to evaluate whether the request should be
granted or not. If the request is permitted heimdall will then execute the 
handler function. If the request is denied than heimdall will call another 
failure handler where you can write out custom messages and status codes as a 
response to the failed request.

### Example policy decision point function

In xacml terms, the policy decision point is the point in the system that 
evaluates access requests against authorization policies before issuing access 
decisions. Heimdall's AuthZFunction is just such a point. This function is 
defined:

	type AuthZHandler func(r *http.Request, token Token, client Client, user User) (status int, message string)

As you can see heimdall will hand you the incoming request, combined with the
token, client, and user. In this function you will then need to return a string
message as well as:

1. heimdall.Permit
2. heimdall.Deny
3. heimdall.Indeterminate
4. heimdall.NotApplicable

Here is an example function:

	pdpFunc(r *http.Request, token heimdall.Token, client heimdall.Client, user heimdall.User) (int, string){
		// In xacml terms heimdall is the 'Policy Enforcement Point' and it has called
		// you. (The 'Policy Decision Point'
		// At this point you should:

		// Contact 'Policy Retrieval Point' and get policy for request
		// If needed you can 'Policy Information Point' and gather additional attributes
		// for the user, client, resource, or anything else

		//Once finished return a policy decision and a message
		if allowed{
			return heimdall.Permit, "Congrats"
		}
		if denied{
			return heimdall.Deny, "You do not have permission"
		}
		if err != nil{
			return heimdall.Indeterminate, err.Error()
		}
		if notsure{
			return heimdall.NotApplicable, "This policy is not applicable to this request"
		}
	}

### Example Failure Function

The failure function allows you to customize the response after you return a 
status other than permit. While it could have been put into the authz function 
we believe it is better to seperate the responsibility into discrete functions.

Here is an example:

	func failFunc(w http.ResponseWriter, r *http.Request, status int, message string, t heimdall.Token, c heimdall.Client, u heimdall.User) {
		if t == nil && (strings.HasSuffix(r.URL.Path, "/") || strings.HasSuffix(r.URL.Path, "/index.html")) {
			http.Redirect(w, r, "/login?return_to="+r.URL.RequestURI(), http.StatusFound)
		} else if t == nil {
			http.Error(w, message, http.StatusUnauthorized)
		} else {
			http.Error(w, message, http.StatusForbidden)
		}
	}

Writing a custom data adapter
---

Obviously the filesystem or in-memory adapter won't work in anything but small 
poc projects. If you need to plug into a database or some other api layer than 
you will want to write your own implementation.

More on custom adapters coming soon!
