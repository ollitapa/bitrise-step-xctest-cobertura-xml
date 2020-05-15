//
//  ExampleTests.swift
//  ExampleTests
//
//  Created by Olli Tapaninen on 15.5.2020.
//  Copyright © 2020 Acme. All rights reserved.
//

import XCTest
@testable import Example

class ExampleTests: XCTestCase {

    func testDoSomeStuff() throws {
        XCTAssertEqual(doSomeStuff(), 42)
    }
}
