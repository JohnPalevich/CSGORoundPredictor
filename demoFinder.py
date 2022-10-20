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
        WebDriverWait(driver, 5).until(lambda s: s.find_element(By.CLASS_NAME, "result-con").is_displayed())
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
    os.chdir(cwd)
    return

def run_go():
    output = subprocess.check_output("go run main.go", shell=True)
    print(output)


if __name__ == "__main__":
    driver = configure_driver() 
    tournament_link = 'https://www.hltv.org/results?event=6372'
    links = find_all_links(driver, tournament_link)
    links_seen = set()
    if os.path.exists("matches_seen.pkl"):
        with open('matches_seen.pkl', 'rb') as f:
            links_seen = pickle.load(f)
    
    # links = {'https://www.hltv.org/matches/2356134/copenhagen-flames-vs-bad-news-eagles-pgl-major-antwerp-2022'}
    links = links - links_seen
    for link in links:
        download_demo(driver, link)
        run_go()
        links_seen.add(link)
        pickle.dump(links_seen, open("matches_seen.pkl", "wb"))
    print("Completed run for tournament:", tournament_link)
    