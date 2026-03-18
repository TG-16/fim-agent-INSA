==========================================================
              WINDOWS FIM AGENT - SETUP GUIDE
==========================================================

This agent monitors file integrity on your Windows system
and ships logs directly to your local OpenSearch instance.

----------------------------------------------------------
1. PREREQUISITES
----------------------------------------------------------
- An active OpenSearch instance.
- Administrator privileges on this PC.
- The 'config.json' must remain in the same folder as the .exe.

----------------------------------------------------------
2. CONFIGURATION
----------------------------------------------------------
Open 'config.json' in a text editor (like Notepad) and 
adjust the following:

- "opensearch_url":  The URL of your OpenSearch (e.g., https://localhost:9200/_bulk)
- "opensearch_user": Your username (default is admin)
- "opensearch_pass": Your password
- "log_level": 1 If you want the agent to log errors only, 2 If you want everything to be logged
- "watch_paths":     A list of folders you want to monitor.
                     Example: ["C:\\Users\\Name\\Desktop"]

----------------------------------------------------------
3. HOW TO RUN
----------------------------------------------------------
A, Right-click 'run_agent.bat'.
B, Select "Run as Administrator".
C, A terminal will open. If you see "FIM Agent Active", 
   the agent is successfully monitoring your files!

----------------------------------------------------------
5. TROUBLESHOOTING
----------------------------------------------------------
- "Access Denied": Ensure you ran the .bat as Administrator.
- "Network Error": Check if OpenSearch is running and your 
                   URL/Credentials in config.json are correct.
- "Deadlock/Crash": Do not watch the entire C:\ drive. 
                    Stick to specific critical folders.
==========================================================