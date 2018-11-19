# Databox Youtube History Driver


Data is gathered using [youtube-dl](https://github.com/rg3/youtube-dl/blob/master/README.md#readme), a shell based program to download youtube videos and their metadata.

In order for the driver to work, the user needs to input their youtube credentials, which will then be passed to the youtube-dl program in order to obtain the data.

When the metadata is downloaded it is then converted into json format, stripped of unnecessary data and stored into the Time Series blob store (TSBlob). 
Data currently is refreshed every 30 seconds (for testing purposes).

## Data stores
The driver has two data stores, one to store user authentication and the other to store video data.
### Credential Store
The authentication data store is a key-value store **(KVStore)** which holds the users username and password. This allows the user to loging with saved credentials if they wish. The content type that is stored inside the store is text **(ContentTypeText)**.

The credential store ID is: ***"YoutubeHistoryCred"***

### Metadata store
The video data store is a time series blob store **(TSBlob)**, which holds JSON objects that contain information on the stored videos. The content type that is stored is JSON ***(ContentTypeJSON)***

The format of the json data stored is:
- FullTitle     (The full title of the video)
- Title         (The main title that is displayed)
- AltTitle      (Any alternative title the video has)
- Dislikes      (Dislike count)
- Views         (View count)
- AvgRate       (Average rating value)
- Description   (The description of the video)
- Tags          (Any tags associated with the video)
- Track         (Any soundtrack used in the video)
- ID            (Unique ID)

The video data store ID is: "YoutubeHistory"

## Example data
```
{"fulltitle":"",
  "title":"How Fast Is Red Dead Redemption's Dead Eye?",
  "alt_title":"",
  "dislike_count":40,
  "view_count":105679,
  "average_rating":4.97663545609,
  "description":"You have to be quick on the draw to survive in Red Dead Redemption, which is why Dead Eye is so helpful. But just how helpful? Kyle takes his best shot on this week's Because Science!\n\nGrab your new Because Science merch here: https://shop.nerdist.com/collections/...\n\nGet a 30-day free trial and watch Because Science episodes early on Alpha: https://goo.gl/QPP3AU\n\nSubscribe for more Because Science: http://bit.ly/BecSciSub\n\nMore science: http://nerdist.com/tag/science/\nWatch more Because Science: http://nerdi.st/BecSci\n\nFollow Kyle Hill: https://twitter.com/Sci_Phile\nFollow us on FB: https://www.facebook.com/BecauseScience\nFollow us on Twitter: https://twitter.com/becausescience\nFollow us on Instagram: https://www.instagram.com/becausescience\nFollow Nerdist: https://twitter.com/Nerdist\n\nBecause Science every Thursday.\n\nLearn More:\n•LIMITS OF HUMAN PERFORMANCE: https://www.blurbusters.com/human-reflex-input-lag-and-the-limits-of-human-reaction-time/2/\n• HUMAN REACTION TIME STATISTICS: https://www.humanbenchmark.com/tests/reactiontime/statistics\n• SPEED OF PROCESSING IN THE HUMAN VISUAL SYSTEM: https://www.nature.com/articles/381520a0\n• THE GUNSLINGER EFFECT: https://pdfs.semanticscholar.org/4ac8/e173f9d53d3caed190d1650c0391e2053d79.pdf?_ga=2.19938488.1202866775.1538594507-800753590.1538594507\n• VIDEO GAMES INCREASE REACTION TIME: https://www.ncbi.nlm.nih.gov/pmc/articles/PMC2871325/\n• MENTAL CHRONOMETRY: https://en.wikipedia.org/wiki/Mental_chronometry\n• SPRINT STARTS AND MINIMUM AUDITORY REACTION TIME [PDF]: https://pdfs.semanticscholar.org/b00c/ac7fad75d4cf3f1fdabdf1ce4cb648247bb7.pdf?_ga=2.134537010.2110069424.1538696101-800753590.1538594507\n• REACTION TIME EXPERIMENTS: https://backyardbrains.com/experiments/reactiontime\n• BREAKDOWN OF A FAST DRAW: http://www.fastdraw.org/fd_draw.html\n• A LITERATURE REVIEW ON REACTION TIME: http://www.cognaction.org/cogs105/readings/clemson.rt.pdf",
  "tags":["Nerdist","fvid","red dead redemption","Because Science","Kyle Hill","cowboy","sharpshooting","marksmanship","wild west","westworld"],
  "track":"",
  "id":"QzIuZudzESo"}
```
