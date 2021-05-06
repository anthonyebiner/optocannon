# *RESISTANCE IS FUTILE*

### How to run:

* Install the latest version of golang to your computer (1.16?)
* Run a postgres instance on your computer attatched to port 31311
* Change host in database.go to your local ip (0.0.0.0 most likely)
* Change user/password to whatever you set
* In the public schema, create a table "graphs" with the following columns:
    * id int
    * input text
    * output text
    * score int
    * solved bool
    * size int
    * processing bool
    * pass int
    * bestpass int
    * name text
    * edges int
* Run parse.py's import_dir function on the dir with the inputs
* Modify the settings in main.go
    * passNum is the pass the solver is on. Increment this one by one as you try different settings
    * the others are general simulated annealing settings. See the commented lines for some settings 
      I used for different runs
* Run the solver (go run *.go)
* Run multiple solvers at once to speed things up
* The database will automatically manage the solvers and dispatch unprocessed jobs to each
* Once all the solvers throw an exception (indicating no more files to process), change settings and run again
* You could also run the solver in solver.py, although it's pretty slow and wasn't used outside of testing
* Uncomment the export lines in main.go and run (go run *.go)
![image](https://user-images.githubusercontent.com/45410382/117250180-60c43000-ae32-11eb-9f78-d12d227faf36.png)
