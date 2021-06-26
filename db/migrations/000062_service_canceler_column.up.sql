BEGIN; 
    ALTER TABLE services   
    ADD COLUMN canceller_id INT REFERENCES users(id) NULL;  
COMMIT;