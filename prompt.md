Taking cubectx as inspiration, I need to build the same kind of tool for Claude called claudectx. This will allow me to switch the settings.local.json file and see multiple different named settings.local.json files for different APIs that are compatible with Claude. 

So it we'll be this file below. I'm not too sure if it supports different files you know of the same name or similar names and point number I need to be able to use different API keys different APIs for different providers. 

'.claude/settings.json'

https://github.com/ahmetb/kubectx

Please explore how kubectx works, and how that has the different conflicts and how it points things between them, and see if we can do a similar kind of thing. I'd really like to just take what works in that, use it in this application. 

I need a tool to run as a command-line tool so I can install it using something like Brew or build it locally and just be able to run the command and do similar kinds of things to `kubectx`. Like `kubectx` and then select the different config, add them, delete them, that kind of thing. 

Fairly open as to what technologies to use and what language to build it in. Happy to take recommendations, but if you are giving me recommendations, please tell me why. 