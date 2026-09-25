package com.example.api;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class PersonList {
    private static final Logger logger = LoggerFactory.getLogger(PersonList.class);

    public void log() {
        logger.info("PersonList using slf4j-api pulled in only via the shared convention plugin");
    }
}
