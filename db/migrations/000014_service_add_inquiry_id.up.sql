BEGIN;

ALTER TABLE services
ADD COLUMN inquiry_id INT NOT NULL;

ALTER TABLE services
   ADD CONSTRAINT fk_inquiry_id
   FOREIGN KEY (inquiry_id)
   REFERENCES service_inquiries(id);

COMMIT;
