from bs4 import BeautifulSoup
from selenium import webdriver
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.common.by import By
from selenium.common.exceptions import TimeoutException
import requests
import os
import subprocess
import pickle

def configure_driver():
    # Add additional Options to the webdriver
    chrome_options = Options()
    # add the argument and make the browser Headless.
    chrome_options.add_argument("--headless")
    # Instantiate the Webdriver: Mention the executable path of the webdriver you have downloaded
    # For linux/Mac
    driver = webdriver.Chrome(executable_path='/Users/john/Downloads/chromedriver' ,options = chrome_options)
    # For windows
    # driver = webdriver.Chrome(executable_path="./chromedriver.exe", options = chrome_options)
    return driver

def find_all_links(driver, link):
    driver.get(link)
    
    try:
        WebDriverWait(driver, 10).until(lambda s: s.find_element(By.CLASS_NAME, "result-con").is_displayed())
    except TimeoutException:
        print("TimeoutException: Element not found")
        return None

    soup = BeautifulSoup(driver.page_source, "lxml")
    
    mydivs = soup.find_all("div", {"class": "result-con"})
    links = set()
    for divTag in mydivs:
        aTags = divTag.find_all("a", {"class": "a-reset"})
        for aTag in aTags:
            links.add("https://www.hltv.org" + aTag['href'])
    #print(links)
    return links

def download_demo(driver, link):
    driver.get(link)

    try:
        WebDriverWait(driver, 5)
    except TimeoutException:
        print("TimeoutException: Element not found")
        return None

    soup = BeautifulSoup(driver.page_source, "lxml")
    demoLink = "https://www.hltv.org" + soup.select_one('a:contains("GOTV Demo")')['href']
    print(demoLink)
    r = requests.get(demoLink, allow_redirects=True)
    match_name = link.split('/')[-1]
    cwd = os.getcwd()
    basePlace = os.getcwd() + "/demo/"
    
    open(basePlace+match_name+'.rar', 'wb').write(r.content)
    os.chdir(basePlace)
    os.system("unrar x "+match_name+".rar")
    for file in os.listdir(basePlace):
        if file.endswith('.dem'):
            parts=file.split('.')
            match_name_parts = parts[0].split('-')
            map_name = match_name_parts[-1]
            count = len(match_name_parts)
            tournName = '-'.join(match_name.split('-')[count-1:])
            os.rename(basePlace+file, basePlace+parts[0]+'-'+tournName+'-'+map_name+'.'+parts[1])
    os.chdir(cwd)
    return

def run_go():
    subprocess.check_output("go run parseDemo.go", shell=True)


if __name__ == "__main__":
    driver = configure_driver() 
    tournament_links = {
        "https://www.hltv.org/results?event=6141",
        'https://www.hltv.org/results?event=6372',
        'https://www.hltv.org/results?event=6140',
        'https://www.hltv.org/results?event=6510',
        'https://www.hltv.org/results?event=6345',
        'https://www.hltv.org/results?event=6587',
        'https://www.hltv.org/results?event=6477',
        'https://www.hltv.org/results?event=6138',
        'https://www.hltv.org/results?event=6137',
        'https://www.hltv.org/results?event=6136',
        'https://www.hltv.org/results?event=6334',
        'https://www.hltv.org/results?event=5730',
        'https://www.hltv.org/results?event=5608',
        'https://www.hltv.org/results?event=5607',
        'https://www.hltv.org/results?event=4866',
        'https://www.hltv.org/results?event=5554',
        'https://www.hltv.org/results?event=5469',
        'https://www.hltv.org/results?event=5971',
        'https://www.hltv.org/results?event=5604',
        'https://www.hltv.org/results?event=5219',
        'https://www.hltv.org/results?event=5728',
        'https://www.hltv.org/results?event=5454',
        'https://www.hltv.org/results?event=5553',
        'https://www.hltv.org/results?event=5552',
        'https://www.hltv.org/results?event=5206'
    }
    tournament_links_seen = set()
    if os.path.exists("tournaments_seen.pkl"):
        with open("tournaments_seen.pkl", 'rb') as f:
            tournament_links_seen = pickle.load(f)
    tournament_links = tournament_links-tournament_links_seen
    
    for tournament_link in tournament_links:
        links = find_all_links(driver, tournament_link)
        links_seen = set()
        if os.path.exists("matches_seen.pkl"):
            with open('matches_seen.pkl', 'rb') as f:
                links_seen = pickle.load(f)
                print(len(links_seen))
        
        #links = {'https://www.hltv.org/matches/2356134/copenhagen-flames-vs-bad-news-eagles-pgl-major-antwerp-2022'}
        links = links - links_seen
        for link in links:
            try: 
                download_demo(driver, link)
                run_go()
            except Exception as e: 
                print("Failed to download/run_go, error message was:", e)
            finally:
                links_seen.add(link)
                pickle.dump(links_seen, open("matches_seen.pkl", "wb"))
                print("Completed demo parsing for match:", link)
        print("Completed run for tournament:", tournament_link)
        tournament_links_seen.add(tournament_link)
        pickle.dump(tournament_links_seen, open("tournaments_seen.pkl", "wb"))