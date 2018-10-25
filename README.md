# Youtube History Driver

## Data format
Data is gathered using youtube-dl, a shell based program to download youtube videos and their metadata.
When the metadata is downloaded it is then converted into json format, stripped of unnecessary data and stored into the Time Series blob store (TSBlob).

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
