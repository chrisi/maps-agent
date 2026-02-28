#ifndef _FLIGHT_DATA_C_H
#define _FLIGHT_DATA_C_H

#ifdef __cplusplus
extern "C" {
#endif

#define MAX_RWR_OBJECTS 40

typedef struct {
    float x;                // Ownship North (Ft)
    float y;                // Ownship East (Ft)
    float z;                // Ownship Down (Ft)
    float xDot;             // Ownship North Rate (ft/sec)
    float yDot;             // Ownship East Rate (ft/sec)
    float zDot;             // Ownship Down Rate (ft/sec)
    float alpha;            // Ownship AOA (Degrees)
    float beta;             // Ownship Beta (Degrees)
    float gamma;            // Ownship Gamma (Radians)
    float pitch;            // Ownship Pitch (Radians)
    float roll;             // Ownship Roll (Radians)
    float yaw;              // Ownship Yaw (Radians)
    float mach;             // Ownship Mach number
    float kias;             // Ownship Indicated Airspeed (Knots)
    float vt;               // Ownship True Airspeed (Ft/Sec)
    float gs;               // Ownship Normal Gs
    float windOffset;       // Wind delta to FPM (Radians)
    float nozzlePos;        // Ownship engine nozzle percent open (0-100)
    float internalFuel;     // Ownship internal fuel (Lbs)
    float externalFuel;     // Ownship external fuel (Lbs)
    float fuelFlow;         // Ownship fuel flow (Lbs/Hour)
    float rpm;              // Ownship engine rpm (Percent 0-103)
    float ftit;             // Ownship Forward Turbine Inlet Temp (Degrees C)
    float gearPos;          // Ownship Gear position 0 = up, 1 = down;
    float speedBrake;       // Ownship speed brake position 0 = closed, 1 = 60 Degrees open
    float epuFuel;          // Ownship EPU fuel (Percent 0-100)
    float oilPressure;      // Ownship Oil Pressure (Percent 0-100)
    unsigned int lightBits; // Cockpit Indicator Lights, one bit per bulb.

    float headPitch;    // Head pitch offset from design eye (radians)
    float headRoll;     // Head roll offset from design eye (radians)
    float headYaw;      // Head yaw offset from design eye (radians)

    unsigned int lightBits2; // Cockpit Indicator Lights, one bit per bulb.
    unsigned int lightBits3; // Cockpit Indicator Lights, one bit per bulb.

    float ChaffCount;   // Number of Chaff left
    float FlareCount;   // Number of Flare left

    float NoseGearPos;  // Position of the nose landinggear
    float LeftGearPos;  // Position of the left landinggear
    float RightGearPos; // Position of the right landinggear

    float AdiIlsHorPos; // Position of horizontal ILS bar
    float AdiIlsVerPos; // Position of vertical ILS bar

    int courseState;    // HSI_STA_CRS_STATE
    int headingState;   // HSI_STA_HDG_STATE
    int totalStates;    // HSI_STA_TOTAL_STATES; never set

    float courseDeviation;     // HSI_VAL_CRS_DEVIATION
    float desiredCourse;       // HSI_VAL_DESIRED_CRS
    float distanceToBeacon;    // HSI_VAL_DISTANCE_TO_BEACON
    float bearingToBeacon;     // HSI_VAL_BEARING_TO_BEACON
    float currentHeading;      // HSI_VAL_CURRENT_HEADING
    float desiredHeading;      // HSI_VAL_DESIRED_HEADING
    float deviationLimit;      // HSI_VAL_DEV_LIMIT
    float halfDeviationLimit;  // HSI_VAL_HALF_DEV_LIMIT
    float localizerCourse;     // HSI_VAL_LOCALIZER_CRS
    float airbaseX;            // HSI_VAL_AIRBASE_X
    float airbaseY;            // HSI_VAL_AIRBASE_Y
    float totalValues;         // HSI_VAL_TOTAL_VALUES; never set

    float TrimPitch;  // Value of trim in pitch axis, -0.5 to +0.5
    float TrimRoll;   // Value of trim in roll axis, -0.5 to +0.5
    float TrimYaw;    // Value of trim in yaw axis, -0.5 to +0.5

    unsigned int hsiBits;  // HSI flags

    char DEDLines[5][26];  //25 usable chars
    char Invert[5][26];    //25 usable chars

    char PFLLines[5][26];  //25 usable chars
    char PFLInvert[5][26]; //25 usable chars

    int UFCTChan, AUXTChan;

    int           RwrObjectCount;
    int           RWRsymbol[MAX_RWR_OBJECTS];
    float         bearing[MAX_RWR_OBJECTS];
    unsigned long missileActivity[MAX_RWR_OBJECTS];
    unsigned long missileLaunch[MAX_RWR_OBJECTS];
    unsigned long selected[MAX_RWR_OBJECTS];
    float         lethality[MAX_RWR_OBJECTS];
    unsigned long newDetection[MAX_RWR_OBJECTS];

    float fwd, aft, total;

    int VersionNum;    // Version of FlightData mem area

    float headX;       // Head X offset from design eye (feet)
    float headY;       // Head Y offset from design eye (feet)
    float headZ;       // Head Z offset from design eye (feet)

    int MainPower;     // Main Power switch state, 0=down, 1=middle, 2=up
} FlightData;

#ifdef __cplusplus
}
#endif

#endif // _FLIGHT_DATA_C_H
