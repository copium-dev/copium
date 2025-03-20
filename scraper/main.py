from algoliasearch.search.client import SearchClientSync
from dotenv import load_dotenv

import json
import os
import time
import datetime
import git
import schedule
import asyncio

"""
special thanks to cvrve for the main logic. changes are made to update algolia
instead of sending messages to discord.
"""

# constants; in new seasons we update repo url. but this is assuming cvrve stays active lol
# if not, we need a new source for scraping... will be a pain 
REPO_URL = "https://github.com/cvrve/Summer2025-Internships"
# if we're in prod we aren't running this from the root project directory so dont have to
# specify the scraper/ prefix
env = os.getenv("ENVRIONMENT")
if env == "prod":
    LOCAL_REPO_PATH = "Summer2025-Internships"
else:
    LOCAL_REPO_PATH = "scraper/Summer2025-Internships"
JSON_FILE_PATH = os.path.join(LOCAL_REPO_PATH, ".github", "scripts", "listings.json")

load_dotenv()
ALGOLIA_APP_ID = os.getenv("ALGOLIA_APP_ID")
ALGOLIA_WRITE_API_KEY = os.getenv("ALGOLIA_WRITE_API_KEY")
ALGOLIA_INDEX_NAME = os.getenv("ALGOLIA_INDEX_NAME")

def clone_or_update_repo():
    print("Cloning or updating repository...")
    if os.path.exists(LOCAL_REPO_PATH):
        try:
            repo = git.Repo(LOCAL_REPO_PATH)
            repo.remotes.origin.pull()
            print("Repository updated.")
        except git.exc.InvalidGitRepositoryError:
            os.rmdir(LOCAL_REPO_PATH)  # Remove invalid directory
            git.Repo.clone_from(REPO_URL, LOCAL_REPO_PATH)
            print("Repository cloned fresh.")
    else:
        git.Repo.clone_from(REPO_URL, LOCAL_REPO_PATH)
        print("Repository cloned fresh.")

def read_json():
    print(f"Reading JSON file from {JSON_FILE_PATH}...")
    with open(JSON_FILE_PATH, "r") as file:
        data = json.load(file)
    print(f"JSON file read successfully, {len(data)} items loaded.")
    return data

# add role; actually is not capable of updating roles but that's fine
async def send_message(message, role):
    print(f"Sending message: {message}")

    try:
        role['objectID'] = role['id']
        _client = SearchClientSync(ALGOLIA_APP_ID, ALGOLIA_WRITE_API_KEY)
        _client.save_object(index_name=ALGOLIA_INDEX_NAME, body=role)
        print(f"Successfully sent data for {role['title']} at {role['company_name']}")
    except Exception as e:
        print(f"Error sending data to Algolia: {e}")

# delete inactive roles
async def send_delete(message, role):
    print(f"Deleting role: {message}")
    try:
        if not role.get('id'):
            print("Role does not exist, skip deletion.")
            return
        _client = SearchClientSync(ALGOLIA_APP_ID, ALGOLIA_WRITE_API_KEY)
        _client.delete_object(index_name=ALGOLIA_INDEX_NAME, object_id=role['id'])
        print(f"Successfully deleted role: {role['title']} at {role['company_name']}")
    except Exception as e:
        print(f"Error deleting role: {e}")

def check_for_new_roles():
    print("Checking for new roles...")
    clone_or_update_repo()
    new_data = read_json()

    if os.path.exists("scraper/previous_data.json"):
        with open("scraper/previous_data.json", "r") as file:
            old_data = json.load(file)
        print("Previous data loaded.")
    else:
        old_data = []
        print("No previous data found.")

    new_roles = []
    deactivated_roles = []

    old_roles_dict = {(role["title"], role["company_name"]): role for role in old_data}

    for new_role in new_data:
        old_role = old_roles_dict.get((new_role["title"], new_role["company_name"]))

        if old_role:
            if old_role["active"] and not new_role["active"]:
                deactivated_roles.append(new_role)
                print(
                    f"Role {new_role['title']} at {new_role['company_name']} is now inactive."
                )
        elif new_role["is_visible"] and new_role["active"]:
            new_roles.append(new_role)
            print(f"New role found: {new_role['title']} at {new_role['company_name']}")

    for role in new_roles:
        print(role)
        message = f"New role: {role['title']} at {role['company_name']}"
        asyncio.run(send_message(message, role))

    for role in deactivated_roles:
        message = f"Role {role['title']} at {role['company_name']} is now inactive."
        asyncio.run(send_delete(message, role))

    with open("scraper/previous_data.json", "w") as file:
        json.dump(new_data, file)
    
    print("Updated previous data with new data.")

    if not new_roles and not deactivated_roles:
        print("No updates found.")

# actually this is run in a separate VM from everything else so we can safely keep this long-running instead of
# scheduling with cron or smth
def main():
    print("Running: ", datetime.datetime.now())
    check_for_new_roles()

if __name__ == "__main__":
    # run once 
    print("Scraper started: ", datetime.datetime.now())
    main()

    schedule.every().hour.do(main)

    while True:
        schedule.run_pending()
        time.sleep(1)