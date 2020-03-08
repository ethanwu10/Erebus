################################
Erebus controller client library
################################

This is the Python library for writing robot controllers for the Erebus
simulation platform.

Defining behavior
=================

A controller's behavior is defined as a class inheriting from
:code:`erebus.client.Behavior` and implementing the class's :code:`tick()` method:

.. code-block:: python
    class Behavior(erebus.client.Behavior):
        def tick(self, sensorData, commands):
            # Your code here

This tick function is called periodically (every 32ms), and is where all of the
logic for your controller should go. The definitions for the :code:`sensorData`
and :code:`commands` objects can be found in the :code:`erebus/client/` folder,
and, the definitions should be fairly self-explanatory for their behavior.

Running
=======

Once a behavior class is defined, you need to create a
:code:`erebus.client.Client` object using it, and call its :code:`run()` method
to start up your client controller and connect to a broker. The :code:`Client`
constructor takes the Behavior object and a name for your client as arguments;
the :code:`run()` method takes a single optional argument, the address of the
broker, and will default to a broker running on your local machine on the
default port.
