//
//  ContentView.swift
//  Example
//
//  Created by Olli Tapaninen on 15.5.2020.
//  Copyright © 2020 Acme. All rights reserved.
//

import SwiftUI

func doSomeStuff() -> Int {
    return 42
}

struct ContentView: View {
    var body: some View {
        Text("Hello, World!")
    }
}

struct ContentView_Previews: PreviewProvider {
    static var previews: some View {
        ContentView()
    }
}
