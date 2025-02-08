from algoliasearch.search.client import SearchClientSync
import json
import os
import time
import git
import schedule
import asyncio

'''
special thanks to cvrve for the code. changes are made to work w/ current project
instead of sending msg to discord...
- update algolia, no need to assign unique id (user just searching by attributes not by id)
'''

# Constants
REPO_URL = 'https://github.com/cvrve/Summer2025-Internships'
LOCAL_REPO_PATH = 'Summer2025-Internships'
JSON_FILE_PATH = os.path.join(LOCAL_REPO_PATH, '.github', 'scripts', 'listings.json')
MAX_RETRIES = 3  # Maximum number of retries for failed channels

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
    with open(JSON_FILE_PATH, 'r') as file:
        data = json.load(file)
    print(f"JSON file read successfully, {len(data)} items loaded.")
    return data

# proposed change for algolia indexing. this not need to go thru rabbitmq since there's no real 'load' just a single
# script called like idk every 15 minutes or something.
ALGOLIA_ID = os.getenv('ALGOLIA_APP_ID')
ALGOLIA_KEY = os.getenv('ALGOLIA_API_KEY')
ALGOLIA_INDEX = os.getenv('INDEX_NAME')

class AlgoliaClient:
    def __init__(self, app_id, api_key):
        self.client = SearchClientSync(app_id, api_key)

    def update_role(self, role, index):
        try:
            if role['active'] and role['is_visible']:
                self.client.save_object(index_name=index, body=role)
                print(f"Indexed: {role['title']} at {role['company_name']}")
            else:
                # Remove from Algolia if inactive or not visible
                self.client.delete_object(role['id'])
                # objectID being company + title is not unique enough so lets use id from the data
                print(f"Removed from index: {role['title']} at {role['company_name']}")

        except Exception as e:
            print(f"Error indexing {role['title']} at {role['company_name']}: {e}")

async def send_message(message, role=None):
    print(f"Sending message: {message}")
    print("--------------------------------")
    if role:
        algolia_client = AlgoliaClient(ALGOLIA_ID, ALGOLIA_KEY)
        algolia_client.update_role(role, ALGOLIA_INDEX)

def check_for_new_roles():
    print("Checking for new roles...")
    clone_or_update_repo()
    new_data = read_json()
    
    if os.path.exists('previous_data.json'):
        with open('previous_data.json', 'r') as file:
            old_data = json.load(file)
        print("Previous data loaded.")
    else:
        old_data = []
        print("No previous data found.")

    new_roles = []
    deactivated_roles = []

    old_roles_dict = {(role['title'], role['company_name']): role for role in old_data}

    for new_role in new_data:
        old_role = old_roles_dict.get((new_role['title'], new_role['company_name']))
        
        if old_role:
            if old_role['active'] and not new_role['active']:
                deactivated_roles.append(new_role)
                print(f"Role {new_role['title']} at {new_role['company_name']} is now inactive.")
        elif new_role['is_visible'] and new_role['active']:
            new_roles.append(new_role)
            print(f"New role found: {new_role['title']} at {new_role['company_name']}")

    for role in new_roles:
        message = f"New role: {role['title']} at {role['company_name']}"
        asyncio.run(send_message(message, role))

    for role in deactivated_roles:
        message = f"Role {role['title']} at {role['company_name']} is now inactive."
        asyncio.run(send_message(message, role))

    with open('previous_data.json', 'w') as file:
        json.dump(new_data, file)
    print("Updated previous data with new data.")

    if not new_roles and not deactivated_roles:
        print("No updates found.")

# IMPORTANT: On the VM we should run this on a cron job as to not waste precious memory!!!!!!
# just for local testing it's running every minute
schedule.every(1).minutes.do(check_for_new_roles)

print("Starting process...")
while True:
    schedule.run_pending()
    time.sleep(1)
