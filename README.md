Revel Persona module
====================

To use:

Reference in app.conf:

    module.persona = gopkg.upstandinghackers.com/revel/modules/persona

    [dev]
    persona.audience = http://localhost:9000

    [prod]
    persona.audience = <insert URL here>

And at the top of your `routes`:

    module:persona
	
Then, wire it into your controller

    type AppController {
	    *revel.Controller
		persona.Persona
	}
	
	func (c *AppController) Index() revel.Result {
		if c.UserEmail != nil {
			// logged in user
		}
	}
